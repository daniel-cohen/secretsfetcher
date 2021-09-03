package secrets

// Our manifest strucutre:
type SecretProviderClass struct {
	Provider      string
	SecretObjects []*SecretObject
	Region        string

	//An optional field to specify a substitution character to use when the path separator character (slash on Linux) is used in the file name.
	// If a Secret or parameter name contains the path separator failures will occur when the provider tries to create a mounted file using the name.
	// When not specified the underscore character is used, thus My/Path/Secret will be mounted as My_Path_Secret. This pathTranslation value can either be the string "False" or a single character string. When set to "False", no character substitution is performed.
	//TOOD: validate that this is a single charactr or "False"
	PathTranslation string //An optional field to specify a substitution character to use when the path separator character (slash on Linux) is used in the file name.
}
