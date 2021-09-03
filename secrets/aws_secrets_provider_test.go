package secrets

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

// func TestGetObjectFromS3(t *testing.T) {
// 	cases := []struct {
// 		client func(t *testing.T) mockSecretmanagerClient
// 		bucket string
// 		key	string
// 		expect []byte
// 	}{
// 		{
// 			client: func(t *testing.T) S3GetObjectAPI {
// 				return mockGetObjectAPI(func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
// 					t.Helper()
// 					if params.Bucket == nil {
// 						t.Fatal("expect bucket to not be nil")
// 					}
// 					if e, a := "fooBucket", *params.Bucket; e != a {
// 						t.Errorf("expect %v, got %v", e, a)
// 					}
// 					if params.Key == nil {
// 						t.Fatal("expect key to not be nil")
// 					}
// 					if e, a := "barKey", *params.Key; e != a {
// 						t.Errorf("expect %v, got %v", e, a)
// 					}

// 					return &s3.GetObjectOutput{
// 						Body: ioutil.NopCloser(bytes.NewReader([]byte("this is the body foo bar baz"))),
// 					}, nil
// 				})
// 			},
// 			bucket: "fooBucket",
// 			key:	"barKey",
// 			expect: []byte("this is the body foo bar baz"),
// 		},
// 	}

// 	for i, tt := range cases {
// 		t.Run(strconv.Itoa(i), func(t *testing.T) {
// 			ctx := context.TODO()
// 			content, err := GetObjectFromS3(ctx, tt.client(t), tt.bucket, tt.key)
// 			if err != nil {
// 				t.Fatalf("expect no error, got %v", err)
// 			}
// 			if e, a := tt.expect, content; bytes.Compare(e, a) != 0 {
// 				t.Errorf("expect %v, got %v", e, a)
// 			}
// 		})
// 	}
// }

func Create(t GinkgoTInterface, secretData map[string]string) *AWSSecretsManagerProvider {
	t.Helper()
	logger := zaptest.NewLogger(t)
	mockClient := &mockSecretmanagerClient{
		data: secretData,
	}

	provider := NewAWSSecretsManagerProviderFromClient(mockClient, logger)
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
						&SecretObject{
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
