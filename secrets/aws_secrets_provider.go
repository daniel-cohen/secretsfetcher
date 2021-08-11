package secrets

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/smithy-go"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap"
)

type AWSSecretsManagerProvider struct {
	// implements SecretsProvider
	zl        *zap.Logger
	awsClient *secretsmanager.Client
	cfg       *SecretProviderClass
}

func NewAWSSecretsManagerProvider(cfg *SecretProviderClass, zl *zap.Logger) (*AWSSecretsManagerProvider, error) {
	//Create a Secrets Manager client
	var (
		awsCfg aws.Config
		err    error
	)

	if cfg.Provider != "aws" {
		return nil, fmt.Errorf("provider %s is not allowed. must be aws", cfg.Provider)
	}

	if cfg.Region == "" {
		awsCfg, err = config.LoadDefaultConfig(context.Background(), config.WithClientLogMode(aws.LogRetries|aws.LogRequest))
	} else {
		awsCfg, err = config.LoadDefaultConfig(context.Background(), config.WithRegion(cfg.Region), config.WithClientLogMode(aws.LogRetries|aws.LogRequest))
	}

	if err != nil {
		return nil, err
	}

	svc := secretsmanager.NewFromConfig(awsCfg)

	return &AWSSecretsManagerProvider{
		awsClient: svc,
		cfg:       cfg,
		zl:        zl.With(zap.String("secretsProvider", "aws_secrets_manger")),
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

	result, err := p.awsClient.GetSecretValue(context.Background(), input)
	if err != nil {
		switch ae := err.(type) {
		case smithy.APIError:
			p.zl.Error("failed to get seceret value",
				zap.String("secreId", secretObj.ObjectName),
				zap.String("errorCode", ae.ErrorCode()),
				zap.String("errorFault", ae.ErrorFault().String()),
				zap.Error(err))

		default:
			// Message from an error.
			p.zl.Error("failed to get seceret value",
				zap.String("secreId", secretObj.ObjectName),
				zap.Error(err))
		}

		return nil, err
	}

	// Decrypts secret using the associated KMS CMK.
	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	//var secretString, decodedBinarySecret string
	var secretString string
	if result.SecretString != nil {
		secretString = *result.SecretString

		//TODO: we might want to prevent logging of secrets values even in debug mode:
		p.zl.Debug("successfully got secret string value",
			zap.String("secreId", secretObj.ObjectName),
			zap.String("value", secretString),
		)

	} else {
		decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(result.SecretBinary)))
		len, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, result.SecretBinary)
		if err != nil {
			p.zl.Error("Base64 Decode Error:", zap.Error(err))
			return nil, err
		}
		secretString = string(decodedBinarySecretBytes[:len])

		p.zl.Debug("successfully got secret binary value",
			zap.String("secreId", secretObj.ObjectName),
			zap.String("value", secretString),
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
