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

install the addon from heroku and create [a heroku http deploy-hook](https://devcenter.heroku.com/articles/deploy-hooks#http-post-hook):

```bash
# add the plugin
heroku addons:add my-sha:test
# get the configuration
deploy_url=$(heroku config | grep MY_SHA_DEPLOY_HOOK_URL | awk '{gsub(/MY_SHA_DEPLOY_HOOK_URL:\n+(.+)/, "", $1); print $2}')
# add the deploy-hook
heroku addons:add deployhooks:http --url=$deploy_url
```
