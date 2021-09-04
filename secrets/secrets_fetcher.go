package secrets

type SecretsFetcher interface {
	Fetch() ([]*Secret, error)
}
