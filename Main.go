package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type ParkingLot struct {
	totalSpaces     int
	availableSpaces int
	waitingQueue    chan int
	mutex           sync.Mutex
	exitChan        chan struct{} // Канал для завершення програми
}

func NewParkingLot(totalSpaces int) *ParkingLot {
	return &ParkingLot{
		totalSpaces:     totalSpaces,
		availableSpaces: totalSpaces,
		waitingQueue:    make(chan int, totalSpaces),
		exitChan:        make(chan struct{}),
	}
}

func (pl *ParkingLot) parkCar(carNumber int) {
	pl.mutex.Lock()
	if pl.availableSpaces > 0 {
		pl.availableSpaces--
		fmt.Printf("Car %d parked in space. Available spaces: %d\n", carNumber, pl.availableSpaces)
		pl.mutex.Unlock()
		return
	}
	pl.mutex.Unlock()

	fmt.Printf("Car %d is waiting for a parking space.\n", carNumber)
	select {
	case pl.waitingQueue <- carNumber:
		select {
		case <-time.After(1 * time.Second):
			select {
			case <-pl.waitingQueue:
				fmt.Printf("Car %d couldn't find a parking space and left.\n", carNumber)
			default:
			}
		}
	}
}

func (pl *ParkingLot) leave(carNumber int) {
	pl.mutex.Lock()
	if pl.availableSpaces < pl.totalSpaces {
		pl.availableSpaces++
	}
	fmt.Printf("Car %d left. Available spaces: %d\n", carNumber, pl.availableSpaces)

	if pl.availableSpaces == pl.totalSpaces {
		// Всі парковочні місця вільні, завершуємо програму
		close(pl.exitChan)
	}

	select {
	case <-pl.waitingQueue:
	default:
	}
	pl.mutex.Unlock()
}

func main() {
	parkingLot := NewParkingLot(7)

	carNumber := 1
	totalCars := 100

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	go func() {
		// Очікуємо, поки всі парковочні місця стануть вільними
		<-parkingLot.exitChan
		fmt.Println("All parking spaces are available. Exiting the program.")
	}()

	for carNumber <= totalCars {
		select {
		case <-ticker.C:
			go func(carNumber int) {
				parkingLot.parkCar(carNumber)
				// Simulate car parked for 4-7 seconds
				time.Sleep(time.Duration(rand.Intn(4000)+4000) * time.Millisecond)
				parkingLot.leave(carNumber)
			}(carNumber)
			carNumber++
		}
	}

	// Очікуємо завершення програми
	<-parkingLot.exitChan
}
