# Road Trip Companion

Welcome to the project for the PROG2005 Cloud Technologies course. 

This project is made by Group 2: *implement group name here*.

This is the Readme for The Road Trip Companion, an application which lets a user plan a road trip. 
In the project the focus is on the use of a REST web application created with the use of third party APIs, the
creation and use of webhooks, firestore and docker. It is created with the programming language Golang and then deployed to a server with docker, using
OpenStack.

**http://10.212.141.222:80/rtc/v1/diag is the basis of our endpoints**

**For endpoint documentation see the project [WIKI](https://git.gvk.idi.ntnu.no/MartinIversen/cloudproject/-/wikis/home)**

<h1>Project Report</h1>

<h3>Startup</h3>

Our initial idea was to make an entertainment hub allowing a user to access information about movies, tv-shows, music, and comics. Sadly we encountered a problem when trying to implement the music API in Golang. We decided that a workaround would be time-consuming and decided to change topic. 

<h3>The Road Trip Companion</h3>

We decided on a road trip travel planner, where we give a user information about the route they are going to travel. For instance where the nearest charging stations are, the shortest path from one destination to another, it will calculate a time of departure depending on the weather conditions. It will also provide the user with points of interest if desired. We had to change the "GPS" API of the project from the Statens Vegvesen API to the TomTom API since Statens Vegvesen did not respond to our email and since their login functionality did not work. 

<h3>Technologies</h3>

The technologies that we decided to implement were Open Stack, Firebase, Cloud functions and Docker. We have been able to implement the following functions to our route planner such as finding the route to the users' destination, find EV stations, filling stations, avoid traffic and traffic incidents, describe the weather and points of interest along your route such as restaurants, hotels and road attractions. We were also able to use Webhooks to notify the user when to travel, based on where they wanted to go, and which weather conditions there were. We decided to scrap some functionality like the ability to get continuous updates from the route endpoint, because this would just result in the route API being called an unhealthy amount of times. This endpoint can instead be a basis for some form of GPS implementation in the future. 

<h3>Reflection</h3>

There was a lot of back and forth in the early stages of our project. No one had any specific ideas in mind which resulted in us having a hard time collectively deciding on something. This on top of the problems with different APIs, keys and paywalls. Once we decided to go for the route planner we had problems with the location API we used,  however, we quickly found another location API providing us with accurate coordinates. One of the things the group benefited from was starting the project early, arranging meetings and discussing topics. Our decision to start coding early on, benefited us when trying to find good APIs. When we decided to use an API, we started poking it immediately, sending requests and understanding its structure. This made troubleshooting early on easy, because we were able to find problems with the APIs fast, and find out if they were able to retrieve the data we wanted. 

We held meetings one to two times a week in order to maintain an overview of the project, what people were working on and problems. To maintain a solid project structure, we decided to give each member a git branch to not overwrite eachothers code in addition to separate different aspects of the project and prevent buggy/breaking code in the master branch. We used git issues, and the issue board frequently to document workflow. By using this board each group member could easily check the status of the project.

If we were to do to the project again, we would have used milestones. By using milestones we would have been able to track the overall progression of the project. Thus, it would have been easier to set different time periods on the different issues to when they needed to be complete. This would have given us a better overview of the different parts of the project that needs to be completed. We would try to have more informative meeting summaries where each group member documented what they have been doing since the last session, in addition to defining what parts of the project are missing. Splitting workload properly and clearly defining tasks for each member using our issue tracker is also something we would have done. We also think using test driven development would have benefited us in terms of time, project structure and general scope in the project.

Our biggest struggle throughout the project has been communication in terms of functionality, clearly defining what an endpoint is supposed to do and what it currently does. This prevented us from getting a clear overview of the project, making the final stages of development hectic and somewhat confusing.  

We struggled with webhook implementation, clearly defining what features we wanted it to have, when to invoke and notify. Furthermore, we tried to implement a client in our project which also resulted in some difficulties. Our problem was that in order to get a notification, the user had to refresh the provided URL. We concluded that this was not very user-friendly, and annoying. Therefore, we removed it and instead implemented Slack, now the user gets a Slack notification with messages given that they provide a Slack URL.

Another hard aspect of the project was the problem of limiting API calls. Due to the way we designed our API it had to call the location API in order to get coordinates for almost all of our endpoints. However, we were able to implement caching in the end by creating an additional collection in Firebase containing previously used locations. We wanted to test the webhook invoke and to do so we needed to skip the sleep function which was hard. Another hard aspect of the project was how to deal with personal traces in the code, every group member has their own way of coding. Implementing functions differently, and uses different variable names. Therefore, it can be hard sometimes to understand what other group members have done without sufficent commenting.  

<h3>Learning outcome</h3>

We have become better at collaborating on programming tasks with eachother. The GitLab issues has been a great tool for the communication in our group. By creating issues in Git, labeling and assigning them to group members. We have learned that good group communication is very important in order to be able to finish a project like this one. We also became better at using GitLab's different tools like issues, branching, WIKI implementation, merge requests and labels. 

We found that having a good foundation with hard policies and principles is important for project work, especially in our field. Not having a good grasp of our project will result in a lot of extra work and confusion. Frequently meeting and discussing the state of the project is crucial for delivering a working application. We have also learned a lot in regard to programming, especially the webhooks and Docker implementation. Expanding on our knowledge of webhooks, how to invoke, cache and structure code to use this technology properly. Learning how to implement Docker, using best practice for Docker compose and Dockerfiles has been challenging but nevertheless rewarding. 
 
<h3>Total work hours dedicated to the project cumulatively by the group:</h3>

The total hours the group has worked on the project has been 88 hours, for a more detailed description of workflow see the [Workflow documentation](https://git.gvk.idi.ntnu.no/MartinIversen/cloudproject/-/wikis/Workflow-Documentation) and the [Issue Board](https://git.gvk.idi.ntnu.no/MartinIversen/cloudproject/-/boards). 
