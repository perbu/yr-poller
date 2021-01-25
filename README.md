# yr-poller
Deamon that polls yr.no for data and emits in on 10min basis. Virtual thermometer.



## Usage

Create a locations.json file, something like this:
```
[
  {
    "id": "tryvannstua",
    "lat": 59.9981362,
    "long": 10.6660856
  },
  {
    "id": "skrindo",
    "lat": 60.6605926,
    "long": 8.5740604
  }
]
```

Compile the project:

```
go build -o poller cmd/main.go
```

Run it
```
./poller
```


It even has some built in help
```
Usage of ./poller:
  -api-url string
    	Baseurl for Yr API (default "https://api.met.no/weatherapi")
  -api-version string
    	API version to use. Appended to URL (default "2.0")
  -interval duration
    	How often to emit data (default 10m0s)
  -locationsfile string
    	JSON file containing locations (default "locations.json")
  -user-agent string
    	User-agent to use (default "yr-poller")
```


## Todo
 * Remove the mutex stuff and use channels. Initially I wrote this not unlike a Java program with a shared data structure
   where I lock/unlock. I've removed most of the access to shared data but the observation cache remains shared.
 * Expand of the sensors supported.
 * Add tests to the timestream code - possibly painful to mock
 * timestream lacks testing
 * statushttp lacks testing

