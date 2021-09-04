package aws

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"go.uber.org/zap/zaptest"
)

// Ref: https://aws.github.io/aws-sdk-go-v2/docs/unit-testing/

// Define a mock struct to be used in your unit tests of myFunc.
type mockSecretmanagerClient struct {
	secretsManagerAPI

	data map[string]string
}

func (m *mockSecretmanagerClient) ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	//TODO:
	return nil, nil

	//ret := &secretsmanager.ListSecretsOutput{}
}

func (m *mockSecretmanagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {

	k := params.SecretId

	if k == nil {
		return nil, errors.New("params.SecretId cannot be nil")
	}

	if v, ok := m.data[*k]; ok {
		return &secretsmanager.GetSecretValueOutput{
			// at his point return params.SecretId as the name
			Name:         k,
			SecretString: &v,
		}, nil
	}

	return nil, errors.New("secret not found")

}

func Create(t GinkgoTInterface, secretData map[string]string) *AWSSecretsManagerProvider {
	t.Helper()
	zl := zaptest.NewLogger(t)
	mockClient := &mockSecretmanagerClient{
		data: secretData,
	}

	provider := newAWSSecretsManagerProviderFromClient(mockClient, "fake_region", zl)
	return provider
}

var _ = Describe(`Fetch users test`, func() {
	var (
		provider *AWSSecretsManagerProvider

		mockDataStore map[string]string = map[string]string{
			"secret1":          "value1",
			"secre2":           "value2",
			"some/secret/name": "{\r\n\t\"user\": \"user-1\",\r\n\t\"password\": \"password-1\",\r\n\t\"host\": \"database.example.com\"\r\n}",
		}
	)
	cases := []struct {
		secreObejectName  string
		expectSecretValue string
		expectError       bool
	}{
		{
			secreObejectName:  "secret1",
			expectSecretValue: "value1",
			expectError:       false,
		},

		{
			secreObejectName:  "secre2",
			expectSecretValue: "value2",
			expectError:       false,
		},

		{
			secreObejectName:  "some/secret/name",
			expectSecretValue: "{\r\n\t\"user\": \"user-1\",\r\n\t\"password\": \"password-1\",\r\n\t\"host\": \"database.example.com\"\r\n}",
			expectError:       false,
		},
		{
			secreObejectName: "missingkey",
			expectError:      true,
		},
	}

	BeforeEach(func() {
		// Load the whole folder:
		provider = Create(GinkgoT(), mockDataStore)
	})

	// Test the static user repo (singlton implementation)
	var TestGetSingleSecretValue = func() {
		Context("Test get a single value", func() {
			for _, testCase := range cases {
				It(`It sould not be nil"`, func() {
					s, err := provider.getSecretValue(
						&AwsSecretObject{
							ObjectName: testCase.secreObejectName,
						},
					)

					if testCase.expectError {
						Expect(err).To(HaveOccurred())
						Expect(s, BeNil())
					} else {
						Expect(err).NotTo(HaveOccurred())
						Expect(s, Not(BeNil()))
						Expect(s.Name, BeEquivalentTo(testCase.secreObejectName))
						Expect(s.Content, BeEquivalentTo(testCase.expectSecretValue))
					}

				})
			}

		})
	}

	TestGetSingleSecretValue()
})
