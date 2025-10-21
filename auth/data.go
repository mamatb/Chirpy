package auth

const (
	refreshTokenLength = 32

	authorizationApiKey    = "ApiKey"
	authorizationBearer    = "Bearer"
	empty                  = ""
	errorInvalidAuthApiKey = "invalid Authorization ApiKey header"
	errorInvalidAuthBearer = "invalid Authorization Bearer header"
	errorMissingAuthApiKey = "missing Authorization ApiKey header"
	errorMissingAuthBearer = "missing Authorization Bearer header"
	headerAuthorization    = "Authorization"
	jwtIssuer              = "chirpy"
	space                  = " "

	regexBcrypt = `^\$2a\$10\$[./0-9A-Za-z]{53}$`
	regexJwt    = `^(eyJ[-_0-9A-Za-z]+\.){2}[-_0-9A-Za-z]+$`
)
