# is-your-isp-blocking-you
A tool to check if the ISP is blocking you for any of the Alexa top 1M websites

## Architecture
![images/is-your-isp-blocking-you.png](./images/is-your-isp-blocking-you.png)

## Screenshots
![scan-stats-db](./images/scan-stats-db.png)
## ToDo :
- [ ] Keep unique domains in the list to scan and remove subdomains - Try to get this to max 100k domains
- [ ] Create this as a CLI tool. See bubble tea golang lib.
- [ ] d3.js or some other tool to create a heat map
- [ ] Keep in DB stats for last run, like : 1. Scan Time 2. Domains scanned 3. Accessible, Non-accessible, blocked, connection timed out domains 4. Location 5. ISP 6. Evil or not 7. Time of scan 8. Type of filtering
- [ ] Decide number of goroutines on the basis of internet connection. A low bandwidth connection will get choked and all websites' will get timed out. Also timeout should be decided on this basis
- [ ] Try out bypasses for common techniques. Keep this as an option in the cli tool.
- [ ] Replace http client with retryable http client - https://github.com/hashicorp/go-retryablehttp
- [ ] Optimise DB connections : https://gorm.io/docs/generic_interface.html#Connection-Pool