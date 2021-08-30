package secrets

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/zap"
)

const (
	defaultMaxResults = 50
)

type AWSSecretsManagerProvider struct {
	// implements SecretsProvider
	zl        *zap.Logger
	awsClient *secretsmanager.Client
	//cfg       *SecretProviderClass
}

//func NewAWSSecretsManagerProvider(cfg *SecretProviderClass, zl *zap.Logger) (*AWSSecretsManagerProvider, error) {
func NewAWSSecretsManagerProvider(region string, zl *zap.Logger) (*AWSSecretsManagerProvider, error) {
	//Create a Secrets Manager client
	var (
		awsCfg aws.Config
		err    error
	)

	//TODO:
	// if cfg.Provider != "aws" {
	// 	return nil, fmt.Errorf("provider %s is not allowed. must be aws", cfg.Provider)
	// }

	var aswOptions []func(*config.LoadOptions) error

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

	return &AWSSecretsManagerProvider{
		awsClient: svc,
		//cfg:       cfg,
		zl: zl.With(zap.String("secretsProvider", "aws_secrets_manger")),
	}, nil
}

func (p *AWSSecretsManagerProvider) getSecretValue(secretObj *SecretObject) (*Secret, error) {
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

	return &Secret{
		Name:    *result.Name,
		Content: secretString,
	}, nil
}

func (p *AWSSecretsManagerProvider) FetchSecrets(secretObjs []*SecretObject) ([]*Secret, error) {
	var res []*Secret
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

func (p *AWSSecretsManagerProvider) FetchAllSecrets(secretNamePrefix string, tagFilters map[string]string) ([]*Secret, error) {
	secretObject, err := p.listSecrets(secretNamePrefix, tagFilters)
	if err != nil {
		return nil, err
	}

	p.zl.Debug("listed secrets", zap.Int("secretCount", len(secretObject)))

	return p.FetchSecrets(secretObject)

}

// We will fetch a list of ARNS and construct SecretObject with the latest versions:
// We can set a range of tag filters . E.g. app=api-verifier
// SecretNamePrefix - is mandatory. E.:g secretNamePrefix= api-verifier/
func (p *AWSSecretsManagerProvider) listSecrets(secretNamePrefix string, tagFilters map[string]string) ([]*SecretObject, error) {
	if strings.TrimSpace(secretNamePrefix) == "" {
		return nil, fmt.Errorf("secretNamePrefix cannot be empty")
	}

	//var secretARNs []string
	var nextToken *string

	var secretObjects []*SecretObject

	filters := []types.Filter{{Key: types.FilterNameStringTypeName, Values: []string{secretNamePrefix}}}

	for k, v := range tagFilters {
		filters = append(filters,
			types.Filter{Key: types.FilterNameStringTypeTagKey, Values: []string{k}},
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

			secretObjects = append(secretObjects, &SecretObject{
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
