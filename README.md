# is-your-isp-blocking-you
A tool to check if the ISP is blocking you for any of the Alexa top 1M websites

## Architecture
![images/is-your-isp-blocking-you.png](./images/is-your-isp-blocking-you.png)

## Screenshots
![scan-stats-db](./images/scan-stats-db.png)
## ToDo :
- [ ] Keep unique domains in the list to scan and remove subdomains - Currently 264k unique domains. Takes ~1330 seconds. Try to get this to max 100k domains
- [ ] Option to use restricted domains from lists like : [CitizenLabs/test-lists](https://github.com/citizenlab/test-lists), [Domains Project](https://github.com/tb0hdan/domains) etc
- [ ] Create this as a CLI tool. See bubble tea golang lib.
- [ ] d3.js or some other tool to create a heat map
- [x] Replace http client with retryable http client - https://github.com/hashicorp/go-retryablehttp
- [x] Keep in DB stats for last run, like : 1. Scan Time 2. Domains scanned 3. Accessible, Non-accessible, blocked, connection timed out domains 4. Location 5. ISP 6. Evil or not 7. Time of scan 8. Type of filtering
- [ ] Decide number of goroutines on the basis of internet connection. A low bandwidth connection will get choked and all websites' will get timed out. Also timeout should be decided on this basis. Can use [speedtest-go](https://github.com/showwin/speedtest-go).
- [ ] Try out bypasses for common techniques. Keep this as an option in the cli tool.
- [ ] Optimise DB connections : https://gorm.io/docs/generic_interface.html#Connection-Pool

## ToDo Server :
- [ ] Create a serverless lambda to send data to.
- [ ] Figure out IP and in turn ISP to be inserted into the DB
- [ ] Give user option in CLI tool to send data to their server
- [ ] Open source this server