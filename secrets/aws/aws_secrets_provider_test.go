package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"go.uber.org/zap/zaptest"
)

// Ref: https://aws.github.io/aws-sdk-go-v2/docs/unit-testing/

// Define a mock struct to be used in your unit tests of myFunc.

type MockAwsSecret struct {
	value string
	tags  map[string]string
	arn   string
}

type mockSecretmanagerClient struct {
	secretsManagerAPI

	data map[string]*MockAwsSecret
}

func doesMatchFilters(secretName string, secret *MockAwsSecret, filters []types.Filter) (bool, error) {
	for _, filter := range filters {
		if len(filter.Values) != 1 {
			return false, fmt.Errorf("only one filter value is allowed for %s", filter.Key)
		}

		switch filter.Key {
		case types.FilterNameStringTypeName:
			if !strings.HasPrefix(secretName, filter.Values[0]) {
				return false, nil
			}
		case types.FilterNameStringTypeTagKey:
			// for a prefix match:
			bFoundKey := false
			for k, _ := range secret.tags {
				if strings.HasPrefix(k, filter.Values[0]) {
					bFoundKey = true
					break
				}
			}
			if !bFoundKey {
				return false, nil
			}

		// for an exact match:
		// if _, ok := secret.tags[filter.Values[0]]; !ok {
		// 	return false, nil
		// }

		case types.FilterNameStringTypeTagValue:
			// for a prefix match:
			bFoundValue := false
			for _, v := range secret.tags {
				if strings.HasPrefix(v, filter.Values[0]) {
					bFoundValue = true
					break
				}
			}
			if !bFoundValue {
				return false, nil
			}
		}
	}

	return true, nil
}

func (m *mockSecretmanagerClient) ListSecrets(ctx context.Context, params *secretsmanager.ListSecretsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretsOutput, error) {
	var res []types.SecretListEntry

	for k, v := range m.data {
		match, err := doesMatchFilters(k, v, params.Filters)
		if err != nil {
			return nil, err
		}

		if match {
			res = append(res, types.SecretListEntry{
				ARN:  &v.arn,
				Name: &k,
			})

		}
	}

	// No paging at this point
	return &secretsmanager.ListSecretsOutput{
		SecretList: res,
	}, nil
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
			SecretString: &v.value,
		}, nil
	}

	return nil, errors.New("secret not found")

}

func CreateProvider(t GinkgoTInterface, secretData map[string]*MockAwsSecret) *AWSSecretsManagerProvider {
	t.Helper()
	zl := zaptest.NewLogger(t)
	mockClient := &mockSecretmanagerClient{
		data: secretData,
	}

	provider := newAWSSecretsManagerProviderFromClient(mockClient, "fake_region", zl)
	return provider
}

// var _ = Describe(`Fetch secret value test`, func() {
// 	var (
// 		provider *AWSSecretsManagerProvider

// 		mockDataStore map[string]*MockAwsSecret = map[string]*MockAwsSecret{
// 			"secret1":          {value: "value1"},
// 			"secre2":           {value: "value2"},
// 			"some/secret/name": {value: "{\r\n\t\"user\": \"user-1\",\r\n\t\"password\": \"password-1\",\r\n\t\"host\": \"database.example.com\"\r\n}"},
// 		}
// 	)
// 	cases := []struct {
// 		secreObejectName  string
// 		expectSecretValue string
// 		expectError       bool
// 	}{
// 		{
// 			secreObejectName:  "secret1",
// 			expectSecretValue: "value1",
// 			expectError:       false,
// 		},

// 		{
// 			secreObejectName:  "secre2",
// 			expectSecretValue: "value2",
// 			expectError:       false,
// 		},

// 		{
// 			secreObejectName:  "some/secret/name",
// 			expectSecretValue: "{\r\n\t\"user\": \"user-1\",\r\n\t\"password\": \"password-1\",\r\n\t\"host\": \"database.example.com\"\r\n}",
// 			expectError:       false,
// 		},
// 		{
// 			secreObejectName: "missingkey",
// 			expectError:      true,
// 		},
// 	}

// 	BeforeEach(func() {
// 		// Load the whole folder:
// 		provider = CreateProvider(GinkgoT(), mockDataStore)
// 	})

// 	// Test the static user repo (singlton implementation)
// 	var TestGetSingleSecretValue = func() {
// 		Context("Test get a single value", func() {
// 			for _, testCase := range cases {
// 				It(`It sould get valid secret"`, func() {
// 					s, err := provider.getSecretValue(
// 						&AwsSecretObject{
// 							ObjectName: testCase.secreObejectName,
// 						},
// 					)

// 					if testCase.expectError {
// 						Expect(err).To(HaveOccurred())
// 						Expect(s, BeNil())
// 					} else {
// 						Expect(err).NotTo(HaveOccurred())
// 						Expect(s, Not(BeNil()))
// 						Expect(s.Name, BeEquivalentTo(testCase.secreObejectName))
// 						Expect(s.Content, BeEquivalentTo(testCase.expectSecretValue))
// 					}

// 				})
// 			}

// 		})
// 	}

// 	//TestGetSingleSecretValue()
// })

// var _ = Describe(`list secrets test`, func() {
// 	var (
// 		provider *AWSSecretsManagerProvider

// 		mockDataStore map[string]*MockAwsSecret = map[string]*MockAwsSecret{
// 			"secret1":                 {value: "value1", arn: "arn1"},
// 			"myprefix":                {value: "value2", arn: "arn2"},
// 			"myprefixABC":             {value: "value3", arn: "arn3"},
// 			"myprefix/with_slash/a/b": {value: "value4", arn: "arn4"},
// 		}
// 	)

// 	cases := []struct {
// 		prefix       string
// 		filters      map[string]string
// 		expectError  bool
// 		expectedArns []string
// 	}{
// 		{
// 			// no prefix
// 			// no filteres
// 			expectError: true,
// 		},
// 		{
// 			// no prefix
// 			filters:     map[string]string{"filter_key": "filter_value"},
// 			expectError: true,
// 		},
// 		{
// 			prefix: "secret1", //exact match
// 			// no filteres
// 			expectError:  false,
// 			expectedArns: []string{"arn1"},
// 		},
// 		{
// 			prefix: "myprefix",
// 			// no filteres
// 			expectError:  false,
// 			expectedArns: []string{"arn2", "arn3", "arn4"},
// 		},

// 		{
// 			prefix: "myprefix/with_slash/",
// 			// no filteres
// 			expectError:  false,
// 			expectedArns: []string{"arn4"},
// 		},
// 	}
// 	fmt.Printf("cases: %v", cases)

// 	BeforeEach(func() {
// 		// Load the whole folder:
// 		provider = CreateProvider(GinkgoT(), mockDataStore)
// 	})

// 	// Test the static user repo (singlton implementation)
// 	var TestListSecrets = func() {
// 		Context("Listing secrets", func() {
// 			for i, testCase := range cases {
// 				It(`It sould list secrets"`, func(i int, tc struct {
// 					prefix       string
// 					filters      map[string]string
// 					expectError  bool
// 					expectedArns []string
// 				}) {

// 					fmt.Printf("testCase: %v\n", testCase)

// 					sos, err := provider.listSecrets(
// 						testCase.prefix,
// 						testCase.filters,
// 					)

// 					if testCase.expectError {
// 						fmt.Printf("Case %d Expecting error\n", i)
// 						Expect(err).To(HaveOccurred())
// 						Expect(sos, BeNil())
// 					} else {
// 						fmt.Printf("Case %d NOT expecting error\n", i)
// 						Expect(err).NotTo(HaveOccurred())
// 						Expect(sos, Not(BeNil()))

// 						// 	Expect(len(sos), BeEquivalentTo(len(testCase.expectedArns)))

// 						// 	//TODO: sort and compare array content
// 					}

// 				})
// 			}

// 		})
// 	}

// 	TestListSecrets()
// })

var _ = Describe("Listing secrets", func() {
	var (
		provider      *AWSSecretsManagerProvider
		mockDataStore map[string]*MockAwsSecret = map[string]*MockAwsSecret{
			"secret1":                 {value: "value1", arn: "arn1"},
			"myprefix":                {value: "value2", arn: "arn2"},
			"myprefixABC":             {value: "value3", arn: "arn3"},
			"myprefix/with_slash/a/b": {value: "value4", arn: "arn4"},

			"matching_prefix/xxx": {value: "value5", arn: "arn5", tags: map[string]string{"sometagnameXXX": "sometagValueYYY"}},
		}
	)
	BeforeEach(func() {
		// Load the whole folder:
		provider = CreateProvider(GinkgoT(), mockDataStore)
	})
	DescribeTable("Listing secret filters",
		func(prefix string,
			filters map[string]string,
			expectError bool,
			expectedArns []string) {

			sos, err := provider.listSecrets(
				prefix,
				filters,
			)

			if expectError {
				Expect(err).To(HaveOccurred())
				Expect(sos).To(BeNil())
			} else {
				Expect(err).NotTo(HaveOccurred())
				//Expect(sos).NotTo(BeNil())

				fmt.Printf("expected_len=%d\n", len(expectedArns))
				fmt.Printf("sos_len=%d\n", len(sos))
				fmt.Printf("expectedArns=%v\n", expectedArns)
				Expect(len(sos)).To(Equal(len(expectedArns)))

				// 	//TODO: sort and compare array content
			}
		},
		Entry("empty prefix, no filters", "", nil, true, nil),
		Entry("empty prefix, with filters", "", map[string]string{"filter_key": "filter_value"}, true, nil),
		Entry("exact name match, no filters", "secret1", nil, false, []string{"arn1"}),
		Entry("prefix matching 3 entries, no iflters", "myprefix", nil, false, []string{"arn2", "arn3", "arn4"}),
		Entry("prefix with a slash, no filters", "myprefix/with_slash/", nil, false, []string{"arn4"}),

		Entry("matching prefix, with filter match both key and value", "matching_prefix", map[string]string{"sometagname": "sometagValue"}, false, []string{"arn5"}),

		Entry("matching prefix, with filter matching key, but not value", "matching_prefix", map[string]string{"sometagname": "sometagValueMISSING"}, false, nil),
	)
})