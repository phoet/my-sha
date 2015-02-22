[MY-SHA](http://addons.heroku.com/my-sha) is an add-on for accessing the current Git SHA of your application.

Most of our applications depend on Git for version control.
For debugging purposes, it is important to know which version is deployed in production.
Because of that, most application developers have it embedded into the Sites HTML-source or the print it out in the logs
when the app is requested with some debugging parameter.

Since Heroku removes all unused files for [slug-compilation](https://devcenter.heroku.com/articles/slug-compiler),
all Git information is lost for the running app.

There are [several workarounds for this problem](http://stackoverflow.com/questions/14583282/heroku-display-hash-of-current-commit), 
ie using a Git-Hook to write the current SHA to some file and check it into version control as well.

This addon uses Herokus built-in deploy-hooks to provide a more streamlined integration.

MY-SHA is accessible via an extremely simple JSON-API and has additional support for ENV variables.

## Provisioning the add-on

MY-SHA can be attached to a Heroku application via the  CLI:

> callout
> A list of all plans available can be found [here](http://addons.heroku.com/my-sha).

```term
$ heroku addons:add my-sha
-----> Adding my-sha to sharp-mountain-4005... done, v18 (free)
```

Once MY-SHA has been added, it exposes all it's information via the `MY_SHA_XXX` environment variables. 
Those settings will be available in the app configuration and will contain the my-sha _app-token_, _plugin-url_, _revision-url_, _deploy-hook-url_ and _revision_.
More on those settings later on.

Those settings can be confirmed using the `heroku config` command.

```term
$ heroku config | grep MY_SHA
MY_SHA_TOKEN:               APP_TOKEN
MY_SHA_URL:                 https://my-sha.herokuapp.com/resources/APP_TOKEN
MY_SHA_REVISION_URL:        https://my-sha.herokuapp.com/revision/APP_TOKEN
MY_SHA_DEPLOY_HOOK_URL:     https://my-sha.herokuapp.com/hook/APP_TOKEN
MY_SHA_REVISION:            JSON_SHA
```

After installing the MY-SHA addon the applications deploy-hook needs to be configured with the deploy-hook-url mentioned above:

```term
$ deploy_url=$(heroku config | grep MY_SHA_DEPLOY_HOOK_URL | awk '{gsub(/MY_SHA_DEPLOY_HOOK_URL:\n+(.+)/, "", $1); print $2}')
$ heroku addons:add deployhooks:http --url=$deploy_url
```

This command creates a new free [a heroku http deploy-hook](https://devcenter.heroku.com/articles/deploy-hooks#http-post-hook) 
on your application with the proper deploy-hook-url of your application.
You can copy and paste the URL from the configuration output as well.

Once everything is set up, _the next deployment_ on heroku will expose your current Git SHA via the revision-url:

```term
$ curl -Ss https://my-sha.herokuapp.com/revision/APP_TOKEN | jq '.'
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

If your plan includes the _Push Git SHA to Environment_ feature, you will be able to get the same information via the `MY_SHA_REVISION` config variable in your applications environment:

```term
$ heroku config | grep MY_SHA_REVISION
MY_SHA_REVISION: {
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

## Rails integration

### Peek-Git

[Peek](https://github.com/peek/peek) users can use a patched [Peek-Git](https://github.com/burn-notice/peek-git) plugin to integrate with the heroku config variables.

Just add it to your Gemfile:

```ruby
# Gemfile
gem "peek-git", github: "burn-notice/peek-git"
```

and enable the plugin in your Peek configuration:

```ruby
# config/initializers/peek.rb
Peek.into Peek::Views::Git
```

## Dashboard

> callout
> For more information on the features available within the MY-SHA dashboard please see the docs at [nofail.de/my-sha/](http://nofail.de/my-sha/).

The MY-SHA dashboard allows you to see detailed installation instructions, your current Git Revision and addon settings.

The dashboard can be accessed via the CLI:

```term
$ heroku addons:open my-sha
Opening my-sha for sharp-mountain-4005...
```

or by visiting the [Heroku Dashboard](https://dashboard.heroku.com/apps) and selecting the application in question. Select MY-SHA from the Add-ons menu.

## Troubleshooting

If there is no Git information showing up on new deployments, make sure that the deploy-hook-url is configured properly.
The deploy-hook can be triggered manually from the deploy-hook-addon page that is accessible through the [Heroku Dashboard](https://dashboard.heroku.com/apps).

## Migrating between plans

> note
> Application owners should carefully manage the migration timing to ensure proper application function during the migration process.

Use the `heroku addons:upgrade` command to migrate to a new plan.

```term
$ heroku addons:upgrade my-sha:pro
-----> Upgrading my-sha:pro to sharp-mountain-4005... done, v18 ($1/mo)
       Your plan has been updated to: my-sha:pro
```

## Removing the add-on

MY-SHA can be removed via the  CLI.

> warning
> This will destroy all associated data and cannot be undone!

```term
$ heroku addons:remove my-sha
-----> Removing my-sha from sharp-mountain-4005... done, v20 (free)
```

## Support

All MY-SHA support and runtime issues should be submitted via one of the [Heroku Support channels](support-channels). 
Any non-support related issues or product feedback is welcome at [nofail.de/my-sha/](http://nofail.de/my-sha/).
