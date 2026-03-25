package graph

// ShortID returns the last 20 characters of a Graph resource id (for display / CLI args).
func ShortID(id string) string {
	if len(id) <= 20 {
		return id
	}
	return id[len(id)-20:]
}
