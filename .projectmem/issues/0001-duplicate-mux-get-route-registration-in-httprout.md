# #0001 Duplicate mux.GET("/") route registration in httprouter.go.tmpl panics at runtime

- 2026-10-24T09:31:57Z `issue`: Duplicate mux.GET("/") route registration in httprouter.go.tmpl panics at runtime [templates/web/cmd/web/routes/httprouter.go.tmpl:56]
- 2026-10-24T09:32:04Z `attempt`: Removed duplicate route registration block and cleaned redundant comments [templates/web/cmd/web/routes/httprouter.go.tmpl:56] (worked)
- 2026-10-24T09:34:59Z `fix`: Removed duplicate mux.GET("/") route registration in httprouter.go.tmpl and cleaned redundant comments across templates
