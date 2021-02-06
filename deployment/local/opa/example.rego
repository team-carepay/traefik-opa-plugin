package example

default allow = false

allow {
	input.method = "GET"
	input.path[0] = "public"
}

allow {
	input.method = "GET"
	input.path = [ "secure", i ]
  has_token([ "123", "456"])
}

has_token(tokens) {
    input.path[1] = tokens[i]
}