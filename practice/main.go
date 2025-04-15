package main

import "fmt"

func main() {
	/** Initialize variable **/
	var card string = "Ace of Spades"
	cards := "Five of Diamonds"

	/*
	* Slice and array
	 */
	cardArray := []string{"Ace", genData()}
	cardArray = append(cardArray, "Spade")

	for i, card := range cards {
		fmt.Println(i, card)
	}

	/*Using Functions for re-assignment*/
	cards = genData()

	fmt.Println(card)
	fmt.Println(cards)
}

func genData() string {
	return "hello"
}
