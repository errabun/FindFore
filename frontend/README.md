# ForeFinder (FindFore frontend)

Legacy Turing capstone README. The monorepo now deploys via **Google Cloud Run** — see repo root `Dockerfile`, `cloudbuild.yaml`, and `CLAUDE.md`.

## Local development

1. `cd frontend`
2. `npm install`
3. `npm run dev` — Vite dev server at http://localhost:3000

Run the Go API separately from the repo root (`go run .`) with `DATABASE_URL` and `JWT_SECRET` set (see `.env.example`).

## Table of Contents

* [Overview](#overview)
* [Installation](#installation)
* [Walkthrough](#walkthrough)

## Overview

ForeFinder is a mobile-focused web application that makes linking up with others for a round of golf easier. Coordination can sometimes be a difficult task when filling spots for a tee time, but ForeFinder provides the tools to make that task a hole in one. Add friends to your ForeFinder network to create a private tee time invitation, or send out an invite to the entire ForeFinder community to get your slots filled ASAP! With future iterations this app will also help courses in securing lost revenue from last minute cancellations by parties, which is a large loss for most courses currently. 

 - Developed in tandem with a team from the back-end program at [Turing School of Software & Design](https://turing.edu/)
 - Built using [React](https://reactjs.org/) + [React Router](reactrouter.com)
 - End-to-end tested with [Cypress](cypress.io)
 - Deployed via Google Cloud Run + Cloud Build (see repo root `cloudbuild.yaml`)
 - Responsive design for mobile, tablet or desktop
 - [PWA](https://developer.mozilla.org/en-US/docs/Web/Progressive_web_apps) optimization for mobile installation + caching


## Installation

### Business:
1. `git clone` this [repo](https://github.com/foreFinder/fore-finder-fe)
2. `cd` into project directory
3. Run `npm i` to install the project dependencies
4. Run `npm start` to begin the development server & open the app in a separate window

### Pleasure:
1. Run the app locally (see Local development above) or deploy to your Cloud Run URL
2. Open the Chrome controls (three dots in the top right) and select "Add to Homepage" (mobile) or "Install ForeFinder" (desktop)
3. The ForeFinder app should now be accessible as an app on your phone's home screen or in your computer's "Chrome Apps" folder
4. You do not need to create a login at this stage of development.  The app and almost all the features are accessible and operational while still in development! 

## Walkthrough

Once ForeFinder loads, the dashboard will be displayed showing a couple widgets, committed tee times and available tee times.  The committed tee times widget displays times that the user has either created or has accepted an invitation to.  Whereas the available tee times can toggle tee time invitations from people in their network of friends, or invitations that are open to the entire ForeFinder community.  

Each of the available tee times show how many positions are still available, the host of the tee time, the actual tee time and the course.  If a user has any available tee times displayed, they will have the option to select Accept, which will add the time to their committed tee times list, or decline, which will delete the tee time from the available list.  Once all the open positions are filled, the tee time will be removed from the available list for all users.

If the current user would like to create a new tee time, they can click the 'Create Tee Time' link in the header.  This will redirect the user to a form to select the course to be played, the date, the tee time, the number of open positions, number of holes and the option to make the event public to the ForeFinder community or private, where they can invite all their friends or specific friends.  Once the form is complete it will notify the user the event has been created and add that event to their committed tee times list. 

The user will also be able to see their friends list and all members of the ForeFinder community from the dashboard (on large screens) or from a link in the header (smaller screens).  Once on the community page, the user will be able to add and delete friends from their community.  


## Demo GIFs

### Accept a Tee Time Invitation

<!-- <img src='' width="400"> -->
![GIF showing an invite accept](./docs/inviteaction.gif)

### Tee Time Event Creation Form

<!-- <img src='' width='400'> -->
![GIF showing event creation](./docs/formsubmission.gif)
                   
### Mobile navigation

![GIF showing mobile app navigation](./docs/mobileNav.gif)
<!-- <img src='' width='400'> -->


Lighthouse Report 

![Lighthouse Report](./docs/lighthouse.png)


## Learning Goals
- Communicating as a full-stack team to ensure data flowed efficently and effectively from front to back
- Learning, developing, and practicing a Git Rebase workflow
- Configuring continuous deployment (originally CircleCI; now Google Cloud Build → Cloud Run)
- Interfacing with PWA technologies to allow mobile installation and caching


## Project Progression

Because this project was our first full-stack effort, planning was an important first step. We used a single GitHub [project board](https://github.com/orgs/foreFinder/projects/1) to add and maintain tasks for both the FE and BE so that we could easily determine what each team was currently working on. Along with building out all applicable user stories, the front-end also utilized a [Miro board](https://miro.com/app/board/o9J_l7sQVZc=/) to wireframe the initial design of the UI and model its components.

Once the planning was complete, the the front-end was able to use the [JSON contract](https://miro.com/app/board/o9J_l7sQVZc=/) modeled by the back-end to build out initial functionality using mock data assuming that the data structure would not change. Once the endpoints were completed and the data we needed skulpted, we were able to link up the back and front-end fairly painlessly.

We continually were referring back to our project board and communicating with one another about progress which was very helpful as we needed to slim down our MVP due to time constraints.  One feature we hoped to have that was later cut out was a login page with authentication of the user.  As we continue to iterate on this project, that will be one feature that we hope to add.  

The last part of the project was to learn a new tech or framework to incorporate into the application.  The front-end team decided to make ForeFinder a progressive web app because we recognized that most users would be using the app from a mobile device.  Seeing as none of the FE team members had experience setting up a PWA, we ran into some struggles with the documentation surrounding service workers and what needed to be included in the manifest.  Luckily, through our continued learned we discovered the framework Workbox which helped clear up the confusion we were facing with writing our own service worker and manifest files. 


## Reflections

All in all, the front-end team was very satisfied with how the project turned out.  We all worked well together as a team, communicated effectively and learned a new feature with how to make an application a functioning PWA.  Seeing as this was the first time that we worked with a back-end team to build a full-stack application, it was very beneficial to learn how to communicate with a back-end team about the FE needs and how we were going to be using/manipulating the data that was being returned from the BE team. Lastly, while it was a bit of a steep learning curve, we all found value with using an interactive git rebase workflow for this project.  

## Authors
<table>
    <tr>
        <td><a href="https://github.com/errabun">Eric Rabun</td>
        <td> <a href="https://github.com/tysnj">Tyson McNutt</td>
        <td><a href="https://github.com/jon-schlandt">Jon Schlandt</td>
        <td><a href="https://github.com/mistercanderson">Chris Anderson</td>
    </tr>
    </tr>
        <td><img src="https://avatars.githubusercontent.com/u/73191225?v=4" alt="E. Rabun" width="125" height="auto" /></td>
        <td><img src="https://avatars.githubusercontent.com/u/65634894?v=4" alt="T McNutt" width="125" height="auto" /></td>
        <td><img src="https://avatars.githubusercontent.com/u/75702270?v=4" alt="J Schlandt" width="125" height="auto" /></td>
        <td><img src="https://avatars.githubusercontent.com/u/73985866?v=4" alt="C Anderson" width="125" height="auto" /></td>
    </tr>
</table>


## Technologies
<table>
    <tr>
        <td>Framework</td>
        <td>Functionality</td>
        <td>Structure</td>
        <td>Styling</td>
        <td>PWA</td>
    </tr>
    </tr>
        <td><img src="https://mildaintrainings.com/wp-content/uploads/2017/11/react-logo.png" alt="react" width="100" height="auto" /></td>
        <td><img src="https://raw.githubusercontent.com/voodootikigod/logo.js/master/js.png" alt="javascript" width="100" height="auto" /></td>
        <td><img src="https://cdn.pixabay.com/photo/2017/08/05/11/16/logo-2582748_1280.png" alt="html" width="100" height="auto" /></td>
        <td><img src="https://www.pinclipart.com/picdir/middle/175-1759459_eng-a-med-kamel-frameworks-css-css-logo.png" alt="css" width="100" height="auto" /></td>
        <td><img src="https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fuser-images.githubusercontent.com%2F3104648%2F28351989-7f68389e-6c4b-11e7-9bf2-e9fcd4977e7a.png&f=1&nofb=1" alt="pwa" width="100" height="auto" /></td>
    </tr>
</table>

