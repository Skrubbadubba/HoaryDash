# Patch: v0.11.5

- Fixed some elements not getting proper theme colors
- Fixed default theme not precalculating all colors

---

## Previous release

## v0.11.0: Themes and a media layout

### Summary

#### ✨ Features

- Named themes
- Media widget:
  - supports playing from context for spotifyplus
  - shuffle/repeat controls
  - can queue via media browser
  - mute toggle
- Theming refactor 
    - Many more semantic color fields
    - Structural overrides for UI elements can now also set:
      - Border radius
      - Border width
      - Padding
- Fullscreen media layout
- New "badge" UI element, currently only used by the fullscreen media layout

#### 🏗 Code changes and optimizations

- Go files are split for better organization
- Use controller pattern to reduce rendered javascript
  - Controllers associate themselves with entityId, returning themselves if same entityId is given
  - Controllers have diffing engine, and notifies listeners only on relevant changes
- Theming utilises per theme based utility classes prefixed with a given css query
  - No entity, layout or other component needs to recieve a theme struct, they style themselves with utility classes and the theming will be applied
- Navigates screens using translateX instead of native scroll, should be slightly more performant

### Release notes (complete wall of text rambling, you have been warned)

Hi to all my 9 users, if any of yall are even here. This has really turned into a passion project, and this is probably the biggest update yet. Been working on it for quite some time, with lots of back and forths, and some breaking changes. So I'd thought I'd share my thoughts. This is just gonna be me yapping, kind of like a devlog.

The theming is by far what took the most time this update. I haven't made a system like this before so I dont really know what I'm doing. But from what I've learned there are two common patterns - semantic and structural fields. Semantic fields are where you define colors such as accents, surfaces and on-surface typography. Basically whaqt we have. Structural fields would instead be where you define values for specific parts of UI elements, so for example saying, tooltips should have this specific border color. Of course these have pros and cons. Semantic coloring means you can sort of "insert" a known pallate, like nord, gruvbox, catpuccin, etc, into names that sound like they would fit, and boom you have a good looking UI. But if you wanted to color some very specific part, you would have to go by trial and error, or look at the source. Structural fields would conversely mean coloring certain parts are easy, but applying a pallette would be cumbersome.

I went with semantic fields, but with the urge to completely overengineer I also wanted the best of both worlds - override the semantic values with optional structural fields. This wasn't impossible or anything but creates a very big tension when coding. Most of all the markup uses these colors directly. But there are natural "UI elements", where those tructural overrides would apply, such as badges, buttons, widgets, etc. Now if these are standalone elements, it would make sense to bundle all of their styling into a complete css class. But ofcourse, the structural overrides are optional. So if we want to assign some of the semantic fields to the elements, we either have to write out all the semantic values we want to use each time we want a UI element, where you could easily forget what values the element was originally composed of. Or you bundle all those values into the element class (".card", ".btn", etc). In that case we are essentially "hijacking" the override system to assign default colors. This is ofcourse not a problem, just that it makes the entire theming architechture consufing from a development perspective.

Now the big thing this update really is the fullscreen media layout. With it comes some other goodies. Since alot of the logic would have been reused between it and the media widget, I decided to extract out a "controller", that handles all communication with HA, syncing state and such. This made adding controls way easier too. So now both the media widget and the media browser got some new features. 

- We now have all the transport controls (shuffle/repeat). We can also toggle mute by clicking the icon. 
- The browser can now also queue items. Ofcourse this relies on the media player supporting it. Now not all items can normally be queued. For example queueing an album failed for my spotifyplus media player. My solution was that if the queue command fails, we call browse on the item instead, getting its children (so songs in an album), and queue all of those. This logic might be brittle though, and I have no idea if its gonna work well on other types of media players.
- As my own use case prmarily is controlling a media player coming from the [spotifyplus](https://github.com/thlucas1/homeassistantcomponent_spotifyplus/tree/master) custom integration, I added an option to "enrich" the controllers behaviour if you know your media player entity is using it. What it does currently is that if you play any item, it tries to play it with the context from where you browsed to it. So for example, browsing to a playlist and choosing a song, we will play that song but in the context of the playlist. Before this, the song just played by itself on repeat, clearing the queue, which made the media browser quite useless. Now unfourtunately there is no autoplay. Playing any item by calling spotifys api doesnt automatically make a queue of suggested songs. Spotify did have an api endpoint you could use to fetch recomended songs based on a seed (such as an album, track, playlist, etc) which would allow use to simulate autoplay, but spotify has deprecated that since some time ago. But I think this is a good enough compromise
  
  Now since the amazing spotifyplus integration pretty much exposes the entire spotify api as HA services, we could technically make a full on spotify client, with the queue visible, search, ability to edit playlist, etc. Thats not in my plans though, just noting that the infrastucture is there.

This pattern has not been implemented for other entities like weather and lights just. I just needed it for the media, and wanted to ship as soon as everything looked good. But I will do that sooner or later. But the plan is to do that with all entities that have more than ultra trivial logic like sensors. The plan is also to use the same pattern, but with "renderers", that way render logic doesnt have to be duplicated each time the entiites of the same type are configured on the dashboard. That should significantly save on the bundled js.

Now some of you may be wondering, what we be regarded as a v1 release? A couple of things I would want before than. 

First, some type of security layer. HoaryDash is never meant to be exposed over the internet anyways, but we could still validate each websocket message in go to make sure the client isnt trying to fetch information about a calendar that was never configured on any dashboard.

Secondly is code cleanup. I have a bunch of these concept of "controllers" and "renderers", but havent implemented them consistently. If I wanted to label this as being v1, I would want to make sure the architechture is completely consistent and readable to other developers.

Thirdly, I would want to have some sort of prefetch step before we build the dashboard, where we communicate with HA. This would be so we get friendly names, icons, etc, from HA before we ever try rendering anything. Here we would also ask HA for some entities, to make a default dashboard with some stuff already populated. I simply thinks its important to setup atleast something that a new user can start with. In my world, throwing someone into a yaml configuration system with nothing to start with is just not the UX a product thats "out of beta" should do.

Lastly, some user feedback. I wouldnt consider "releasing" this unless I have gotten atleast some points of user feedback, be that bugs or suggestions, that I can apply. Now the thing is, all these previous points also relies on user interest. I'm actually pretty fine with where this project is at currently. Feature-wise, it has what I set out to make. I have no need for security. With no eyes from other developers, there isn't really an incentive to cleanup the code. With no users, there is no incentive for new features. That being said, I am seriously looking for users. I think minimizing e-waste is important, and if anyone can have a serious, feature-rich and good looking HA dashboard alternative to lovelace, working on 12+ year old tablets, thats a real incentive to go looking for them.

Have a good one 🫡 