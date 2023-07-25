package lexer

import "github.com/KSpaceer/yayamls/token"

type TokenStream interface {
	Next() token.Token
}
