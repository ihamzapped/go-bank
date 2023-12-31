package main

func main() {
	app := NewApiServer(":8000")
	app.Run()
}
