
# Option Assistant

Option Assistant is a backend service and frontend service designed to help evaluate an investor's current position and the potential impact of option trades. The project is under development, with plans to introduce a React-based frontend for a comprehensive user interface.

Accuracy, performance, and scalability by design are the key measures of success for this project.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Before running the project, ensure you have Go installed on your machine. You can download and install Go from [the official Go website](https://golang.org/dl/).

### Installing

Clone the repository to your local machine:

```
git clone https://github.com/jcdevguru/option-assistant
```

Navigate to the project directory:

```
cd option-assistant
```

Install the dependencies:

```
go mod tidy
```

### Running the Server

The project uses `Taskfile.yml` for managing tasks such as running the server. To start the server, execute the following command:

```
task server:run
```

This command will start the backend server, making the API accessible for requests.  In development, the default port is 8080.

### API Documentation

The API is documented using Swagger, making it easy to understand and interact with the available endpoints. To view the API documentation, navigate to the `docs` folder:

```
cd docs
```

Open the Swagger documentation file in your browser to explore the available API endpoints and their specifications.

## Example

When running locally, the `/optionChain` endpoint can be used to get a matrix of prices calculated via the standard Black-Scholes formula for a Call or Put option, projected over a range of asset prices, strike prices, and days to expiry.  Here is a sample call done with `cUrl`:

```sh
curl -X 'GET' \
  'http://localhost:8080/optionChain?assetName=ACME&optionType=Call&assetPriceLow=130&assetPriceHigh=140&assetPriceStep=1&strikePriceLow=125&strikePriceHigh=150&strikePriceStep=1&daysToExpiryLow=1&daysToExpiryHigh=90&daysToExpiryStep=1&riskFreeRate=0.1&volatility=0.2' \
  -H 'accept: application/json'
```

A freshly restart server running locally on a Macbook Air M2 ran the above in 50 ms.

The endpoint accepts query arguments as follows:


| Query Token         | Argument | Function                                                                                      |
|---------------------|----------|-----------------------------------------------------------------------------------------------|
| `assetName`         | ACME     | Specifies the name of the asset for which the options chain is requested.                     |
| `optionType`        | Call     | Determines the type of option (Call or Put) in the options chain.                             |
| `assetPriceLow`     | 130      | Sets the lower bound for the asset price range.                                               |
| `assetPriceHigh`    | 140      | Sets the upper bound for the asset price range.                                               |
| `assetPriceStep`    | 1        | Defines the step value for iterating through asset prices within the specified range.         |
| `strikePriceLow`    | 125      | Sets the lower bound for the strike price range.                                              |
| `strikePriceHigh`   | 150      | Sets the upper bound for the strike price range.                                              |
| `strikePriceStep`   | 1        | Defines the step value for iterating through strike prices within the specified range.        |
| `daysToExpiryLow`   | 1        | Sets the lower bound for the days to expiry range.                                            |
| `daysToExpiryHigh`  | 90       | Sets the upper bound for the days to expiry range.                                            |
| `daysToExpiryStep`  | 1        | Defines the step value for iterating through days to expiry within the specified range.       |
| `riskFreeRate`      | 0.1      | Specifies the risk-free interest rate used in options pricing models.                         |
| `volatility`        | 0.2      | Specifies the volatility of the asset's returns, used in options pricing models.              |

## Upcoming Features

This project is in its WIP stages and is not yet ready for release.  

- **React-based Frontend:** An intuitive and responsive frontend using React is in development. This will provide a user-friendly interface to interact with the backend services.

## Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
