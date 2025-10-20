package auth

const (
	HeaderAuthorization    = "Authorization"
	JwtIssuer              = "chirpy"
	ErrorMissingAuthBearer = "missing Authorization Bearer header"
	ErrorInvalidAuthBearer = "invalid Authorization Bearer header"
	ErrorMissingAuthApiKey = "missing Authorization ApiKey header"
	ErrorInvalidAuthApiKey = "invalid Authorization ApiKey header"
)
