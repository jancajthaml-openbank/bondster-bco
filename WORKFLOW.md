# Retrieve LOGIN scenario

```
POST https://bondster.com/ib/proxy/router/api/public/authentication/getLoginScenario
```


# Validate LOGIN scenario

```
POST https://bondster.com/ib/proxy/router/api/public/authentication/validateLoginStep

{
  "scenarioCode": "USR_PWD",
  "authProcessStepValues": [
    {
      "authDetailType": "USERNAME",
      "value": "xxx"
    },
    {
      "authDetailType": "PWD",
      "value": "xxx"
    }
  ]
}

```

# Get Transaction IDS list for given time range

```
POST https://bondster.com/ib/proxy/mktinvestor/api/private/transaction/search

{
  "valueDateFrom": {
    "month": "1",
    "year": "2018"
  },
  "valueDateTo": {
    "month": "12",
    "year": "2018"
  }
}

```

# Get Transactions details given transaction IDS

```
POST https://bondster.com/ib/proxy/mktinvestor/api/private/transaction/list

{
  "transactionIds": []
}
```
