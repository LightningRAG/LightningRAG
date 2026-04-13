package i18n

import "fmt"

// LocaleError carries an i18n key (and optional format args) so that the API
// layer can translate the message per request locale instead of hard-coding it.
type LocaleError struct {
	Key  string
	Args []any
}

func (e *LocaleError) Error() string {
	if len(e.Args) > 0 {
		return fmt.Sprintf(T(DefaultLocale, e.Key), e.Args...)
	}
	return T(DefaultLocale, e.Key)
}

// NewError creates an error that carries an i18n key.
// Usage:  return i18n.NewError("svc.user.username_taken")
func NewError(key string) error {
	return &LocaleError{Key: key}
}

// NewErrorf creates an error that carries an i18n key with format arguments.
// Usage:  return i18n.NewErrorf("svc.kb.unsupported_format", ext)
func NewErrorf(key string, args ...any) error {
	return &LocaleError{Key: key, Args: args}
}

// TranslateError resolves a LocaleError to the target locale.
// For plain errors it returns err.Error() unchanged.
func TranslateError(locale string, err error) string {
	if err == nil {
		return ""
	}
	if le, ok := err.(*LocaleError); ok {
		if len(le.Args) > 0 {
			return Tf(locale, le.Key, le.Args...)
		}
		return T(locale, le.Key)
	}
	return err.Error()
}
