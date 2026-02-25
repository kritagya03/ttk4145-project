# ttk4145-project

TTK4145 elevator project at NTNU, made with Joachim Frydenlund and Henrik Markestad :)

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
1. Implement initialization (e.g. initialize between floors)
2. Implement marking call orders
3. Implement marking call completed
4. Implement arriving at floor
5. Implement door timeout
6. Implement checking if has requests above, here, below
7. Implement choosing direction
8. Implement if should stop
9. Implement turning on and off lights
10. Implement setting motor
11. Implement setting floor indicator
12. Implement acceptance tests
13. Implement door obstruction
14. Implement motor stuck

Watchdog:
1. make a process pair instead

Hardware:
1. Implement reciever
2. Implement transmitter
3. Implement server

Models:
1. Set correct durations for all the Time.Duration
2. Better code quality for writing models


Better code quality for importing models