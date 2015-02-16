require 'net/http'
require 'uri'
require 'json'
require 'pp'

Net::HTTP.post_form(URI(ENV['MY_SHA_DEPLOY_HOOK_URL']), app: 'my-sha', user: 'phoet', url: 'http://localhost:5000', head: 'eaf6069', head_long: 'eaf6069', prev_head: '', git_log: ' * phoet: log the body', release: 'v123')

response = Net::HTTP.get(URI(ENV['MY_SHA_REVISION_URL']))
pp JSON.parse(response)
