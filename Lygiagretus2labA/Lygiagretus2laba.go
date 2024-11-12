package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type BasketBallerData struct {
	Id               int     `json:"id"`
	LastName         string  `json:"lastName"`
	BirthYear        int     `json:"age"`
	PointsPerGame    float64 `json:"pointsPerGame"`
	PrimeNumberCount int
}

type InputDataManager struct {
	dataChannel    chan BasketBallerData
	commandChannel chan Command
}
type OutputDataManager struct {
	filteredDataChannel chan BasketBallerData
	commandChannel      chan Command
}
type Command struct {
	action string
	player *BasketBallerData
}

func NewOutputDataManager(maxSize int) *OutputDataManager {
	return &OutputDataManager{
		filteredDataChannel: make(chan BasketBallerData),
		commandChannel:      make(chan Command, maxSize*3),
	}
}
func NewInputDataManager(maxSize int) *InputDataManager {
	return &InputDataManager{
		dataChannel:    make(chan BasketBallerData, maxSize),
		commandChannel: make(chan Command, maxSize*3),
	}
}

func (dm *InputDataManager) AddPlayer(player BasketBallerData) {
	dm.dataChannel <- player

}
func (dm *InputDataManager) RemovePlayer(player BasketBallerData) {
	fmt.Printf("Announcing removal of player with ID %d: %s\n", player.Id, player.LastName)
}
func writeResultsToFile(filteredPlayers chan BasketBallerData, filename string, done chan struct{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	for player := range filteredPlayers {
		_, err := fmt.Fprintf(file, "ID: %d, LastName: %s, BirthYear: %d, PointsPerGame: %.2f, PrimeNumberCount: %d\n",
			player.Id, player.LastName, player.BirthYear, player.PointsPerGame, player.PrimeNumberCount)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}
	}
	done <- struct{}{}
}
func (dm *InputDataManager) ProcessInputData() {
	for command := range dm.commandChannel {
		switch command.action {
		case "add":
			if command.player != nil {
				dm.AddPlayer(*command.player)
			}
		case "remove":
			if command.player != nil {
				dm.RemovePlayer(*command.player)
			}
		case "shutdown":
			close(dm.dataChannel)

		}
	}
}
func (dm *OutputDataManager) Filter(player BasketBallerData) {
	if player.PrimeNumberCount >= 295 && player.PointsPerGame > 15 {
		dm.filteredDataChannel <- player
	}
}
func (dm *OutputDataManager) ProcessResultsData() {
	for command := range dm.commandChannel {
		switch command.action {
		case "filter":
			if command.player != nil {
				dm.Filter(*command.player)
			}
		case "shutdown":
			close(dm.filteredDataChannel)
			return
		}
	}
}
func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n == 2 || n == 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

func performHeavyCalculation(number int) int {

	start := time.Now()
	primeCount := 0
	for {
		if isPrime(number) {
			primeCount++
		}
		number--

		if time.Since(start) >= 20*time.Millisecond {
			break
		}
	}
	return primeCount
}
func worker(workerID int, dataChannel chan BasketBallerData, doneChannel chan struct{}, cmd chan Command, output OutputDataManager) {
	for player := range dataChannel {
		fmt.Printf("Announcing add of player with ID %d: %s\n", player.Id, player.LastName)

		player.PrimeNumberCount = performHeavyCalculation(player.BirthYear)
		filterCommand := Command{"filter", &player}
		output.commandChannel <- filterCommand
		removeCommand := Command{"remove", &player}
		cmd <- removeCommand

	}
	doneChannel <- struct{}{}
}

func main() {
	startTime := time.Now()

	file, err := os.ReadFile("data3.json")
	if err != nil {
		fmt.Println("Error reading JSON file:", err)
		return
	}

	var players []BasketBallerData
	if err := json.Unmarshal(file, &players); err != nil {
		fmt.Println("Error parsing JSON data:", err)
		return
	}

	maxSize := len(players) / 2
	inputDataManager := NewInputDataManager(maxSize)
	outputDataManager := NewOutputDataManager(maxSize)

	go inputDataManager.ProcessInputData()
	go outputDataManager.ProcessResultsData()

	numWorkers := 4
	doneChannel := make(chan struct{}, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go worker(i, inputDataManager.dataChannel, doneChannel, inputDataManager.commandChannel, *outputDataManager)
	}
	doneFile := make(chan struct{})
	go writeResultsToFile(outputDataManager.filteredDataChannel, "filtered_players.txt", doneFile)

	for _, player := range players {
		inputDataManager.commandChannel <- Command{"add", &player}
	}

	inputDataManager.commandChannel <- Command{"shutdown", nil}

	for i := 0; i < numWorkers; i++ {
		<-doneChannel
	}
	outputDataManager.commandChannel <- Command{"shutdown", nil}
	time.Sleep(1 * time.Millisecond)
	elapsedTime := time.Since(startTime)
	fmt.Printf("Length of data channel: %d\n", len(inputDataManager.dataChannel))
	fmt.Printf("All workers finished processing in %s.\n", elapsedTime)

}
