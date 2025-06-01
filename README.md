# ct-log-monitor

ct-log-monitor is a simple certificate transparency (CT) log monitoring application.

It can be used to monitor CT logs for potentially malicious domains.

#### It does the following:
- Fetches a list of active and qualified CT logs
- Writes the name and URL of these logs to `config/logbook.json`
- Writes the Tree Size / Signed Tree Head (STH) to `config/logbook.json` for each log
- Writes the previous STH to `config/logbook.json` from previous runs of the application
- Spins up goroutines for each CT log that has had new log entries
- Gets all certs and their cert.Subject.CommonName (the primary domain name / host the cert is issued for) and their DNSNames
- Checks these values against a series of strings defined in `config/monitor_domains.json`

# Requirements
Golang

# Usage
Create a `monitor_domains.json` file in the `config/` directory with the following structure:

#### Note: Values in `monitor_domains.json` cannot have spaces in them because of strict string comparison used in `similarity.go`
```
[
    {
        "brand": "acompany",
        "domain": "acompany.com",
        "aliases": [
            "a-company",
            "companya"
        ]
    }
]
```

Open up a terminal in the project root directory and run the following command:
```
go run main.go
```

This will run the application and create and/or update the `config/` directory with the following files:

- `logbook.json`: A list of all CT Logs and their STHs

- `logged_domains.json`: A list of domains discovered in CT logs that are similar to those defined in `monitor_domains.json`

# example_data
The `example_data/` directory contains some examples of structures for the following files:
- `logbook.json`
- `logged_domains.json`
- `monitor_domains.json`

While `logbook.json` and `logged_domains.json` will automatically be created when running the application, monitor_domains.json needs to be created before execution.

# TODOs
- Tests
- Improved string comparison
