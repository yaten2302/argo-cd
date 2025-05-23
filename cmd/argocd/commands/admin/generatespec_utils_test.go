package admin

import (
	"bytes"
	"os"
	"testing"

	utilio "github.com/argoproj/argo-cd/v3/util/io"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetOutWriter_InlineOff(t *testing.T) {
	out, closer, err := getOutWriter(false, "")
	require.NoError(t, err)
	defer utilio.Close(closer)

	assert.Equal(t, os.Stdout, out)
}

func TestGetOutWriter_InlineOn(t *testing.T) {
	tmpFile := t.TempDir()
	defer func() {
		_ = os.Remove(tmpFile + ".back")
	}()

	out, closer, err := getOutWriter(true, tmpFile)
	require.NoError(t, err)
	defer utilio.Close(closer)

	assert.Equal(t, tmpFile, out.(*os.File).Name())
	_, err = os.Stat(tmpFile + ".back")
	require.NoError(t, err, "Back file must be created")
}

func TestPrintResources_Secret_YAML(t *testing.T) {
	out := bytes.Buffer{}
	err := PrintResources("yaml", &out, &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "my-secret"},
		Data:       map[string][]byte{"my-secret-key": []byte("my-secret-data")},
	})
	require.NoError(t, err)

	assert.Equal(t, `apiVersion: v1
kind: Secret
metadata:
  name: my-secret
stringData:
  my-secret-key: my-secret-data
`, out.String())
}
