# littleApiTestingTool
 A small testing tool for internal API testing. Ment to simulate a client sending infromations on a system through POSTS.
 This is being built as my personal project with a very specific use case in my company, so it focuses specifically on generating POST requests to a dynamically generated endpoint.
 While the endpoint can indeed be dynamically generated, and the generation could be used to define arbitrary endpoints, the tool is ment for a very specifically formatted endpoint:
  'https://ADDRESS:PORT/API_ROUTE/{DYNAMIC_ENDPOINT}'
 Currently, as previously stated, the program only manages POST requests since that was my needed use case, but this can easily changed in the code to manage different HTTP requests.
 This is because my use case specifically requires only POST requests, so I thought it unnecessary to add more burden on the configuration of messages.
 That being said I will probably add other types of HTTP reqests shortly.

## Build
Pre-built binaries can be downloaded from the Releases tab, otherwhise they can be built as follows.
 Requirements:
  -The golang compiler
 '''bash
 git clone https://github.com/PalmieriClaudio/littleApiTestingTool.git
 cd littleApiTestingTool
 go build init.go
 '''
 this will generate an executable file.
 The configuration files found in the folder with the executable are necessary to configure the endpoints and requests sent to them.
 Currently those need to be setup in the same folder as the executable, this will be changed in the future.
 No installation or external dependencies are necessary.

 For more specific build configurations (like building for a different OS o to a specific instal folder), specific build configurations can be found  at https://go.dev/doc/tutorial/compile-install

## Usage
NOTE: The actual software is still being built so this is changing daily while I figure out how I would like it to work.
At startup the program will offer a few options:
 - reload configuration. This will reload the config.json file in case any changes where done during the runtime of the program.
 - Send messages. This will send the static messages configured in the data.yaml. These are ment for one time testing of functionality and are sent once.
 - Start simulation. This will send the messages in sim.yaml. These can be configured to change dynamically and are sent concurrently on a timer defined in the sim.yaml file.

Both data.yaml and sim.yaml come with an exemple file associated with every line commented to exemplify how they need to be configured.

The simulation messages are all sent together at the start, and then sent concurrently on the timer defined as 'frequency'.
In order to avoid race conditions between the messages, a dependency can be set that will allow the message to be sent only if the messages it depends on have also been sent previously. (this is yet to be implemented)
Variables can be set in the messages, and these come in different types and have different behaviours.
Most importantly some types are consumed while other are not.
Consumable types:
 - sequence
 - range
 These types are consumable in the sense that they define a list of values that will be cycled over, and when all of the values have been used the message that used them will stop sending.
Non consumable types:
 - static
 - random
These will have a value defined at runtime which will statically be sent for all the messages.
