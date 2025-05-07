package fofa

type queryOptions struct {
	fields []string
}

type WithQueryOption func(*queryOptions)

func WithExtraFields(fields ...string) WithQueryOption {
	return func(o *queryOptions) {
		o.fields = append(o.fields, fields...)
	}
}
