package api

type Context struct {
	Server  string
	Session string
}

func InitContext(server, session string) *Context {
	return &Context{
		Server:  server,
		Session: session,
	}
}
