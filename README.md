# FTP Mass Anonymous Login Checker

This is a simple Go program to perform anonymous FTP login to a list of hosts, retrieve their geo-location information, and store the results in a file.

## Requirements

- Go programming language
- Dependencies:
  - `github.com/jlaffaye/ftp`

## Usage

1. Clone the repository:
git clone https://github.com/Cyb3rWarCrim1nal/FTP-Mass-Anonymous-Login-Checker.git

2. Install dependencies: go mod init github.com/Cyb3rWarCrim1nal/FTP-Mass-Anonymous-Login-Checker; go mod tidy

3. Create a `hosts.txt` file containing a list of hostnames or IP addresses, with each entry on a separate line.

4. Run the program: go run ftpscanner.go



The program will perform anonymous FTP login to each host in the `hosts.txt` file, retrieve their geo-location information using the ip-api.com API, and store the results in a file named `found.txt`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
