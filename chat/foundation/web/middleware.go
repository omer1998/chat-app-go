package web

//represent a handler function designed to run code before or after
// another handler
type MidFunc func(handler HandlerFunc) HandlerFunc

// wrapMiddleware creates a new handler by wrapping middleware around a final
// handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func WrapMiddleWare(mw []MidFunc, handler HandlerFunc) HandlerFunc {
	// TODO: does it work right
	// do we need to loop in backward way in order to ensure tha the first middleware run firest
	for i := 0; i < len(mw); i++ {
		mwFunc := mw[i]
		if mwFunc != nil {
			handler = mw[i](handler)

		}

	}
	return handler
}
