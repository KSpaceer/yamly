package lexer

import "github.com/KSpaceer/fastyaml/token"

type TokenStream interface {
	Next() token.Token
}
