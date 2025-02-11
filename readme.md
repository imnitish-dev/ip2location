# IP2Location gRPC and REST api Service

This project is a Rest and gRPC-based IP geolocation lookup service that integrates with MaxMind and IP2Location databases to fetch location details based on an IP address.

## Features
- Fetch geolocation details using MaxMind and IP2Location databases.
- Supports gRPC for high-performance communication.
- REST API available for easier integration.
- Automatic database updates.

## Installation

### Prerequisites
Ensure you have the following installed:
- Go (1.18+)
- MaxMind GeoLite2 account
- IP2Location LITE account


### Clone the Repository
```sh
git clone git@github.com:imnitish-dev/ip2location.git
cd ip2location
```

### Environment Configuration
Copy the example environment file and update it with your credentials:
```sh
cp .env.example .env
```
Modify `.env` with your credentials:
```ini
# Server Configuration
PORT=3011
HOST=0.0.0.0
GRPC_PORT=50051

# MaxMind Configuration
MAXMIND_ACCOUNT="your_account"
MAXMIND_LICENSE_KEY="your_license_key"

# IP2Location Configuration
IP2LOCATION_TOKEN="your_token"
IP2LOCATION_CODE="your_code"
```

## Usage

### Start the Service
Run the service with:
```sh
go run main.go
```

### Using gRPC
Invoke the gRPC service on port `50051`.
Example request:
```json
{
  "ip": "xx.xx.xx.xx"
}
```

### Using REST API
Endpoint:
```sh
GET https://ip2locapi.imnitish.dev/lookup/<ip>
```
```sh
GET https://ip2locapi.imnitish.dev
```
Example response:
```json
{
  "maxmind": {
    "country": "India",
    "city": "Mumbai",
    "region": "Maharashtra",
    "latitude": 0.0,
    "longitude": 0.0,
    "country_code": "IN"
  },
  "ip2location": {
    "country": "India",
    "city": "Mumbai",
    "region": "Maharashtra",
    "latitude": 0.0,
    "longitude": 0.0,
    "country_code": "IN"
  }
}
```

## Updating the Database
To update the database daily, use the provided script:
```sh
./update-db.sh
```
Automate the process by adding a cron job:
```sh
0 0 * * * /path/to/update-db.sh
```


## Author
[imnitish-dev](https://github.com/imnitish-dev)

