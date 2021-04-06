package static

var (
	// ApplicationMode represents the mode the application runs in
	// PROD makes the application run in production mode, every other value forces it to run in debug mode
	ApplicationMode = "DEV"
	Production      = ApplicationMode == "PROD"

	// ApplicationVersion represents the version string to display
	ApplicationVersion = "DEV-localbuild"

	// KarenServiceName represents the service name to use for this service when reporting incidents to karen
	KarenServiceName = "canalization"

	// DomainsRedisKey represents the Redis key under which all valid domains are saved
	// As this key should not change in any time we just force it here
	DomainsRedisKey = "__domains"
)
