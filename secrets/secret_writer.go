package secrets

import (
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/zap"
)

// We can have secrets written to different outputs. E.g: to std out, files, etc..
type SecretWriter interface {
	// slashConversionChar- if the secret name contains slashes, we'll convert them to this char:
	WriteSecrets(secretRes []*Secret)
}

type FileSecretWriter struct {
	zl                  *zap.Logger
	outputFolder        string
	stopOnWriteError    bool
	slashConversionChar string
}

// StopOnError - the the stopOnWriteError flag which will cause the writer to stop write errors.
func (sfw *FileSecretWriter) StopOnError() *FileSecretWriter {
	sfw.stopOnWriteError = true
	return sfw
}

func NewFileSecretWriter(
	outputFolder string,
	slashConversionChar string,
	zl *zap.Logger) *FileSecretWriter {
	return &FileSecretWriter{
		zl:                  zl,
		outputFolder:        outputFolder,
		slashConversionChar: slashConversionChar,
	}
}

// writeSecrets - writes the secrets to disk. If outputFolder is empty we'll output to the current folder
// slashConversionChar - all slashes win the secret name will be replaced with this char.
// If empty, no replacement will occur (can fail on wrie to the disk)
func (sw *FileSecretWriter) WriteSecrets(secretRes []*Secret) error {

	// pathTranslationChar := sw.awsCfg.PathTranslation
	// if slashConversionChar != "" {
	// 	pathTranslationChar = slashConversionChar
	// }

	var result *multierror.Error
	for _, v := range secretRes {
		// outputFileName := v.Name
		// if pathTranslationChar != "" {
		// 	if pathTranslationChar != pathTranslationFalse {
		// 		outputFileName = strings.ReplaceAll(outputFileName, "/", pathTranslationChar)
		// 	}
		// } else {
		// 	outputFileName = strings.ReplaceAll(outputFileName, "/", defaultPathTranslation)

		// }

		outputFileName := v.Name
		if sw.slashConversionChar != "" {
			outputFileName = strings.ReplaceAll(outputFileName, "/", sw.slashConversionChar)
		}

		outputFilePath := path.Join(sw.outputFolder, outputFileName)

		sw.zl.Info("writing secret to file",
			zap.String("file_path", outputFilePath),
			zap.String("secret_name", v.Name),
		)
		if err := ioutil.WriteFile(outputFilePath, []byte(v.Content), os.ModePerm); err != nil {
			sw.zl.Error("failed to write file", zap.String("file_path", outputFilePath), zap.Error(err))
			if sw.stopOnWriteError {
				return err
			}

			// Skip it and move on
			// record the erro in the multi error and move on:
			result = multierror.Append(result, err)
			continue
		}

	}

	// Returning a multierror only if there are errors
	return result.ErrorOrNil()
}
