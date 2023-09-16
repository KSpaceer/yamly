package encode

import (
	"bytes"
	"github.com/KSpaceer/yamly"
	"gopkg.in/yaml.v3"
	"io"
	"strings"
)

var _ yamly.TreeWriter[*yaml.Node] = (*ASTWriter)(nil)

type ASTWriter struct{}

func (*ASTWriter) WriteTo(dst io.Writer, tree *yaml.Node) error {
	return yaml.NewEncoder(dst).Encode(tree)
}

func (w *ASTWriter) WriteString(tree *yaml.Node) (string, error) {
	var sb strings.Builder
	if err := w.WriteTo(&sb, tree); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func (w *ASTWriter) WriteBytes(tree *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	if err := w.WriteTo(&buf, tree); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
