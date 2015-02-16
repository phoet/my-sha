# my-sha

get git information about your repository

```bash
curl -Ss https://my-sha.herokuapp.com/revision/YOUR_API_TOKEN | jq '.'
{
  "app": "my-sha",
  "user": "phoet",
  "url": "http://my-sha.herokuapp.com",
  "head": "4f52cea",
  "prev_head": "",
  "head_long": "4f52cea",
  "git_log": "",
  "release": "v26"
}
```

## usage

install the addon from heroku

TODO (show how to do it and get the deploy-hook url)

you gotta have [a heroku http deploy-hook](https://devcenter.heroku.com/articles/deploy-hooks#http-post-hook) setup that points to your my-sha endpoint:

```bash
heroku addons:add deployhooks:http --url=https://my-sha.herokuapp.com/hook/YOUR_API_TOKEN
```
