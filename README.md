# ttk4145-project

TTK4145 elevator project at NTNU, made with Joachim Frydenlund and Henrik Markestad :)

## How to run

```sh
elevatorserver

sudo apt update && sudo apt install gnome-terminal
git clone https://github.com/kritagya03/ttk4145-project.git
cd ./ttk4145-project/cmd/elevator
go run ./main.go -network-id 1
```

Testing:
```sh
simelevatorserver --port 16002

cd ./ttk4145-project/cmd/elevator
go run ./main.go -network-id 2 -hardware-port 16002

simelevatorserver --port 16003

cd ./ttk4145-project/cmd/elevator
go run ./main.go -network-id 3 -hardware-port 16003
```

Or install the project through go

## Ask studass
* If behavior is idle and door is closed, can it go to another floor before opening doors
* Does the obstruction sensor being on imply door open?
* When starting with door open between floors, the elevator moves with the door open. Is this allowed?
* If elevator moves between floors to serve call, but receives powerloss between floors, should it continue in the same direction on startup?

## TODO

Master:
1. Combine MasterWorldviews--
2. Implement SlaveTimeout by reassigning orders--
3. Make applyMasterState()--
4. Test Master Server-
5. Implement acceptance tests for checking the slaves
5. Make sure possible for three masters to combine at the same time on NewMasterConnection
6. Go through comments
7. Better code quality-
    - Don't pass floorcount, elevatorcount and buttontypes. Use matrix dimentions

Slave:
1. Implement initialization (e.g. initialize between floors)-
2. Implement marking call orders-
3. Implement marking call completed-
4. Implement arriving at floor-
5. Implement door open timeout-
6. Implement checking if has requests above, here, below-
7. Implement choosing direction-
8. Implement if should stop-
9. Implement turning on and off lights-
10. Implement setting motor-
11. Implement setting floor indicator-
12. Implement door obstruction-
13. Implement acceptance tests
14. Implement motor stuck (acceptance test)

Watchdog:
1. make a process pair instead

Hardware:
1. Implement reciever (maybe not currently working)
2. Implement transmitter (maybe not currently working)
3. Implement server

Models:
1. Set correct durations for all the Time.Duration
2. Better code quality for writing models


Better code quality for importing models

Maybe switch from printing to logging.

Fault Tolerance:
process pairs
go through specs
acceptance tests




Current errors:
 Poll rate in 10ms, heartbeat is 5 ??
* when master dies most of the slaves die (but one might become the new master on rare occasions)
* When any elevator adds a hall call on the same floor as a elevator, that elevator also turns on its cab call light on that floor
* When a slave is in floor 0, slave has cab calls on floors 1,2,3, slave moving between floor 0 to floor 1, master dies, a cab order on floor 0 spawns.


 Possible huge erros:
 * We are passing [][]CallState slices around believing we get deep copies, but a slice is maybe represented as a pointer such that we get shallow copies. We might need to implement a deep copy function for [][]CallState. Since SlaveWorldview and MasterWorldview uses [][]CallState we might need to implement a deep copy for both worldviews too.

Possible rrrors:
