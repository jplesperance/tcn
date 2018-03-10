package main

// Meat of the app
func main() {
	bc := NewBlockchain()

	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
