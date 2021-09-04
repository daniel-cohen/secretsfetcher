package aws

type AwsSecretObject struct {
	ObjectName    string
	ObjectVersion string
	//ObjectType    string // - currently we will only support the secretsmanager object type
	//objectAlias
	ObjectVersionLabel string // object version stage, default to latest if empty
}
