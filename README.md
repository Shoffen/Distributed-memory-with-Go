# Distributed-memory-with-Go
Task for parallel programming course at university

# Basketball Player Data Processing

This Go application processes a JSON file containing basketball player data, applying parallel programming techniques to perform heavy computations and filtering based on specified criteria.

## Overview

The program is designed to read player data from a `data.json` file and perform the following tasks:

- Add players to a data processing queue.
- Calculate the count of prime numbers based on each player's birth year.
- Filter players based on their points per game and prime number count.
- Write the filtered results to a text file.

## Key Features

- **Concurrent Processing**: Utilizes multiple goroutines to handle data processing in parallel, improving efficiency and speed.
- **Dynamic Channel Management**: Implements channels for communication between the main function, input manager, output manager, and worker goroutines.
- **Filtering Logic**: Filters players based on their performance metrics, ensuring only relevant data is saved.
