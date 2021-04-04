package static

var (
	// ApplicationMode represents the mode the application runs in
	// PROD makes the application run in production mode, every other value forces it to run in debug mode
	ApplicationMode = "DEV"

	// ApplicationVersion represents the version string to display
	ApplicationVersion = "DEV-localbuild"
)
