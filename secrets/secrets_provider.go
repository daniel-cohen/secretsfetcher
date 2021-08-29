package secrets

// Use this code snippet in your app.
// If you need more information about configurations or implementing the sample code, visit the AWS docs:
// https://docs.aws.amazon.com/sdk-for-go/v1/developer-guide/setting-up.html

type Secret struct {
	Name    string
	Content string
}

//TODO:
// One implementation can a local file.
// type SecretsProvider interface {
// 	FetchSecrets() ([]*Secret, error)
// }

// type SecretProviderConfig interface {
// 	//Get()
// }
