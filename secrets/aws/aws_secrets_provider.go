package aws

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/smithy-go"
	"github.com/daniel-cohen/secretsfetcher/logging"
	"github.com/daniel-cohen/secretsfetcher/secrets"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/zap"
)

const (
	defaultMaxResults = 50
)

type secretsManagerAPI interface {
	ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error)
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

type AWSSecretsManagerProvider struct {
	zl        *zap.Logger
	awsClient secretsManagerAPI
	region    string
}

func newAWSSecretsManagerProviderFromClient(awsClient secretsManagerAPI, region string, zl *zap.Logger) *AWSSecretsManagerProvider {
	//Create a Secrets Manager client
	return &AWSSecretsManagerProvider{
		awsClient: awsClient,
		zl:        zl.With(zap.String("secretsProvider", "aws_secrets_manger")),
		region:    region,
	}
}

func NewAWSSecretsManagerProvider(region string, zl *zap.Logger) (*AWSSecretsManagerProvider, error) {
	//Create a Secrets Manager client
	var (
		awsCfg aws.Config
		err    error
	)

	awsLogger := logging.NewAwsLogger(zl)
	aswOptions := []func(*config.LoadOptions) error{config.WithLogger(awsLogger)}

	// Enable aws debug logging:
	if zl.Core().Enabled(zap.DebugLevel) {
		aswOptions = append(aswOptions, config.WithClientLogMode(aws.LogRetries|aws.LogRequest))
	}

	if region != "" {
		aswOptions = append(aswOptions, config.WithRegion(region))
	}

	awsCfg, err = config.LoadDefaultConfig(context.Background(), aswOptions...)
	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(awsCfg)

	return newAWSSecretsManagerProviderFromClient(svc, region, zl), nil
}

func (p *AWSSecretsManagerProvider) Region() string {
	return p.region
}

func (p *AWSSecretsManagerProvider) getSecretValue(secretObj *AwsSecretObject) (*secrets.Secret, error) {
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretObj.ObjectName), // this can be the name or full ARN
	}

	if secretObj.ObjectVersion != "" {
		input.VersionId = aws.String(secretObj.ObjectVersion)
	}

	if secretObj.ObjectVersionLabel != "" {
		input.VersionStage = aws.String(secretObj.ObjectVersionLabel)
	}

	logFields := []zap.Field{
		zap.String("objectName", secretObj.ObjectName),
		zap.String("objectVersion", secretObj.ObjectVersion),
		zap.String("objectVersionLabel", secretObj.ObjectVersionLabel),
	}

	result, err := p.awsClient.GetSecretValue(context.Background(), input)
	if err != nil {
		switch ae := err.(type) {
		case smithy.APIError:
			p.zl.With(logFields...).Error("failed to get seceret value",
				zap.String("errorCode", ae.ErrorCode()),
				zap.String("errorFault", ae.ErrorFault().String()),
				zap.Error(err))

		default:
			// Message from an error.
			p.zl.With(logFields...).Error("failed to get seceret value", zap.Error(err))
		}

		return nil, err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString

		p.zl.With(logFields...).Debug("successfully got secret string value",
			zap.Stringp("secretArn", result.ARN),
		)

	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			p.zl.With(logFields...).Error("Base64 Decode Error:", zap.Error(err))
			return nil, err
		}
		secretString = string(decodedBinarySecretBytes[:len])

		p.zl.With(logFields...).Debug("successfully got secret binary value",
			zap.Stringp("secretArn", result.ARN),
		)
	}

	return &secrets.Secret{
		Name:    *result.Name,
		Content: secretString,
	}, nil
}

func (p *AWSSecretsManagerProvider) FetchSecrets(secretObjs []*AwsSecretObject) ([]*secrets.Secret, error) {
	var res []*secrets.Secret
	// Get the values one by one:
	for _, secretObj := range secretObjs {
		if secret, err := p.getSecretValue(secretObj); err != nil {
			// It will be logged and we'll continue to other secrets, we don't want to stop:
			continue
		} else {
			res = append(res, secret)
		}
	}

	return res, nil
}

func (p *AWSSecretsManagerProvider) FetchAllSecrets(secretNamePrefix string, tagKeyFilters []string, tagValueFilters []string) ([]*secrets.Secret, error) {
	secretObjects, err := p.listSecrets(secretNamePrefix, tagKeyFilters, tagValueFilters)
	if err != nil {
		return nil, err
	}

	p.zl.Debug("listed secrets", zap.Int("secretCount", len(secretObjects)))

	return p.FetchSecrets(secretObjects)

}

// We will fetch a list of ARNS and construct AwsSecretObject with the latest versions:
// We can set a range of tag filters . E.g. app=api-verifier
// SecretNamePrefix - is mandatory. E.:g secretNamePrefix= api-verifier/
func (p *AWSSecretsManagerProvider) listSecrets(secretNamePrefix string, tagKeyFilters []string, tagValueFilters []string) ([]*AwsSecretObject, error) {
	if strings.TrimSpace(secretNamePrefix) == "" {
		return nil, fmt.Errorf("secretNamePrefix cannot be empty")
	}

	//var secretARNs []string
	var nextToken *string

	var secretObjects []*AwsSecretObject

	filters := []types.Filter{{Key: types.FilterNameStringTypeName, Values: []string{secretNamePrefix}}}

	for _, v := range tagKeyFilters {
		filters = append(filters,
			types.Filter{Key: types.FilterNameStringTypeTagKey, Values: []string{v}},
		)
	}

	for _, v := range tagValueFilters {
		filters = append(filters,
			types.Filter{Key: types.FilterNameStringTypeTagValue, Values: []string{v}},
		)
	}

	// do while we have more secrets to page through:
	for {
		output, err := p.awsClient.ListSecrets(context.Background(), &secretsmanager.ListSecretsInput{
			Filters:    filters,
			MaxResults: defaultMaxResults,
			NextToken:  nextToken,
		})

		if err != nil {
			p.zl.Error("request to list secretes failed", zap.Error(err))
			return nil, err
		}

		for _, secret := range output.SecretList {
			p.zl.Info("secret listed",
				zap.Stringp("arn", secret.ARN),
				zap.Stringp("name", secret.Name),
				zap.Any("tags", secret.Tags),
			)

			if secret.ARN == nil {
				return nil, fmt.Errorf("recieved empty ARN")
			}

			secretObjects = append(secretObjects, &AwsSecretObject{
				ObjectName: *secret.ARN,
			})
		}

		if output.NextToken == nil {
			break
		}

		nextToken = output.NextToken
	}

	return secretObjects, nil
}
