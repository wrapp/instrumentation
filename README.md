# Instrumentation

This repository contains the instrumentation library used in various service to monitor their behaviour

## Client
Here's an example of setting a simple `GET /user/1` client.

```
    c, err := client.New()
	if err != nil {
		// do something...
	}
	resp, err := c.Get(context.Background(), "http://localhost/user/1",
		client.UserAgent("my-user-agent"))
	if err != nil {
		// do something ...
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))

	// Output:
	// {"id": 42, "email": "user@meniga.com"}
```
