# secretsfetcher

A CLI tool to fetch secrets from a secret management store and write them to the file system as files.

The aws command will fetch multiple secrets and write them as individual files to the output folder

Usage:
  secretsfetcher aws [flags]

Flags:
* -h, --help                  help for aws
* -m, --manifest {manifest_filepath}       secrets manifest file
* -o, --output {folder_path}         output folder (writes multiple json files to this folder)
* --prefix string           a prefix for all secrets to fetch
* --tagkeys stringArray     an array of tag key prefixes of filters to find secerts by. Example: --tagkeys=app,secret-type
* --tagvalues stringArray   an array of tag value prefixes of filters to find secerts by. Example: --tagvalues=my-app-name,b44c6886-96c4-4b4d-b267-30d7c5787b1a


## Configuration
We allow configuration in 3 ways leveraging viper (in priority order)

1. config file
2. ENV vars
3. commandline arguments (where available)



Sample environment variable override values:

```
"APP_LOGLEVEL": "debug",
"APP_AWS_PREFIXFILTER": "my-app/user-secrest/",
"APP_AWS_PATHTRANSLATION": "@",
"APP_AWS_REGION": "ap-southeast-2"

"APP_AWS_TAGKEYFILTERS": "app,user-type",
"APP_AWS_TAGVALUEFILTERS": "my-app,some-id",
```

Sample configuration file:

  ```yaml
LogLevel: info

Aws:
  prefixFilter: "mysecretprefix/"
  tagKeyFilters:
    - tag_name_prefix1
    - tag_name_prefix2
  tagValueFilters:
    - tag_value_prefix1
  Region: ""
  PathTranslation: "_"
  

## Operation modes

The aws secrets fetcher command can operate in 2 modes:

1. Using a secrets manifest file.  
    Note: Make sure the IAM role policy allows:
 
    ```json
    "Action": "secretsmanager:GetSecretValue",
    ```

2. Listing and fetching all secrets using a prefix + tag (key/value) filters.  
    Make sure your secrets are labeled properly and that the IAM role policy allows:

    ```json
    "Action": "secretsmanager:ListSecrets"
    ```

Sample IAM policy to allow both modes:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "VisualEditor0",
            "Effect": "Allow",
            "Action": "secretsmanager:GetSecretValue",
            "Resource": "arn:aws:secretsmanager:us-west-2:111122223333:secret:SOME_SECRET_PREFIX/*"
        },
        {
            "Sid": "VisualEditor1",
            "Effect": "Allow",
            "Action": "secretsmanager:ListSecrets",
            "Resource": "*"
        }
    ]
}
```


### Mode 1: Secrets manifest

The secrets manifest file lists secrets to fetch by secretsfetcher.
The secrets manifest has a similar structure to the aws SecretProviderClass CRD. As desribed here:
* [https://github.com/aws/secrets-store-csi-driver-provider-aws#secretproviderclass-options](https://github.com/aws/secrets-store-csi-driver-provider-aws#secretproviderclass-options)
* [https://aws.amazon.com/blogs/security/how-to-use-aws-secrets-configuration-provider-with-kubernetes-secrets-store-csi-driver/](https://aws.amazon.com/blogs/security/how-to-use-aws-secrets-configuration-provider-with-kubernetes-secrets-store-csi-driver/)
* https://secrets-store-csi-driver.sigs.k8s.io/getting-started/usage.html

paramters:
* pathTranslation: An optional field to specify a substitution character to use when the path separator character (slash on Linux) is used in the file name. If a Secret or parameter name contains the path separator failures will occur when the provider tries to create a mounted file using the name. When not specified the underscore character is used, thus My/Path/Secret will be mounted as My_Path_Secret. This pathTranslation value can either be the string "False" or a single character string. When set to "False", no character substitution is performed.
* region: An optional field to specify the AWS region to use when retrieving secrets from Secrets Manager or Parameter Store. If this field is missing, the provider will lookup the region from the annotation on the node. This lookup adds overhead to mount requests so clusters using large numbers of pods will benefit from providing the region here.


Sample manifest file: 
```yaml
provider: aws
secretObjects:
  - objectName: "arn:aws:secretsmanager:us-west-2:111122223333:secret:aes128-1a2b3c"
    objectType: "secretsmanager"
    objectVersion: "ab24b1be-c0a9-4b07-841d-cd9df6f480e9"  # [OPTIONAL] object version id, default to latest if empty
  - objectName: "MySecret2"
    objectType: "secretsmanager" 
    objectVersionLabel: "AWSCURRENT"  # [OPTIONAL] object version stage, default to latest if empty
  - objectName: "MySecret3"
    objectType: "secretsmanager" 


region: ap-southeast-2
pathTranslation: "$"
```

For comparison, this is a equivalent aws SecretProviderClass :

```yaml
apiVersion: secrets-store.csi.x-k8s.io/v1alpha1
kind: SecretProviderClass
metadata:
  name: aws-secrets
spec:
  provider: aws
  parameters:
    objects: |
      array:
        - |
          objectName: "arn:aws:secretsmanager:us-west-2:111122223333:secret:aes128-1a2b3c"
          objectType: "secretsmanager"
          objectVersion: "ab24b1be-c0a9-4b07-841d-cd9df6f480e9"
        - |
          objectName: "MySecret2"
          objectType: "secretsmanager"
          objectVersionLabel: "AWSCURRENT"
        - |
          objectName: "MySecret3"
          objectType: "secretsmanager"
   region: ap-southeast-2
   pathTranslation: "$"
```


### Mode 2: List Secrets (search) and fetch them all

You can configure for following search parameters (supported through cli flags, configuration and ENV vars):
1. prefixFilter - Will list all secrets with that name prefix (this is a wildcard search).
2. tagKeyFilters- A list of (prefix) AWS tag key name to filter for (a match must include tags with all prefixes)
3. tagValueFilters - A list of (prefix) AWS tag values to filter for (a match must include tags with all prefixes)


