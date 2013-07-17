# "Puppet" by High Fidelity

Puppet is an experimental IMU-to-browser interface using websockets.

## Theory of Operation

1. A web browser (the "client") loads a client-side javascript application from Puppet's built in webserver.
2. The client requests configuration data, and starts the application, which establishes a websocket connection to the server.
3. Accelerometer, gyroscope, and magnetometer data is captured from an Invensense MPU-9150 nine-DOF MEMS sensor. 
4. The captured data is fed through an implementation of Madgwick's IMU/AHRS algorithm, which integrates the sensor data, and outputs a quaternion position estimate. 
5. The quaternion estimate is streamed via an integrated websocket server to the client, which renders the sensor's orientation in 3D.

## Goals

* 50Hz rendering updates, including transit time.

## Quick Start

1. Ensure that you have Go installed on your machine. You may either [download a pre-built binary](https://code.google.com/p/go/downloads/list) for your platform, or (if you're on OS X) install it via [Homebrew](http://mxcl.github.io/homebrew/) using the following command:

        $ brew install go

2. Clone the `hifi-puppet` repository using Git:

        $ git clone https://git.github.com/kenkeiter/hifi-puppet

3. Build the executable for your platform. The makefile will automatically download and install any dependencies -- but it may take a second. Literally.

        $ cd hifi-puppet/
        $ make

4. Start the server:

        $ export GOMAXPROCS=4
        $ build/hifi-puppet --imu_port /dev/tty.<insert port here>

  You should see Puppet attempt to connect to the attached sensor board. Once this is successful, you may proceed to step 5.  

5. You're done! Open a webkit-compatible browser, and direct it to: [http://localhost:8192/](http://localhost:8192/)

## Configuration

### Command-Line Options

* `--host <host>` -- set the host and port upon which the interface will be served. This typically defaults to `0.0.0.0:8192`, and should follow a similar format.
* `--asset_path <path>` -- specifies a specific absolute path from which interface assets will be served. This currently defaults to `/var/lib/hifi-puppet`, which is where assets are installed by default.

### Connecting from Outside Your LAN

In order to allow users to connect to Puppet from outside your LAN, you'll need to do the following:

1. Disable the firewall on your computer, or be sure to allow access to port 8192 via TCP.
2. Enable port-forwarding on your router. You'll need to forward port 8192 from your WAN to your local machine.
3. Determine your public IP address (you can [find out from Google](https://www.google.com/search?q=whats+my+ip&oq=whats+my+ip)), and direct remote clients to `http://<your public IP>:8192/`.

## Development

You can build and test Puppet using the instructions in the "Quick Start" section. When working on the web interface, you'll need to have the `sass` gem installed, as well as the CoffeeScript compiler. Google for how to get these set up.

Then, to watch the appropriate folders for changes:

      $ cd src/www
      $ scss -w src/scss/:styles/
      $ coffee -o scripts/ -cw src/coffeescripts/

## Known Issues

* Occasionally, the sensor interface will not be reset to its previous state when the application is exited; when restarted, the application will hang while connecting to the sensor interface. If this occurs, simply unplug the board, plug it back in, and attempt to start the application again.
* There's not a lot of error handling; if things go wrong, a thread may crash -- but typically the application will stay up.

## References

1. [MPU9150Lib](https://github.com/Pansenti/MPU9150Lib/blob/master/libraries/MPU9150Lib/MPU9150Lib.cpp) -- A common library for the MPU-9150
2. [SerialInterface.cpp](https://github.com/worklist/hifi/blob/master/interface/src/SerialInterface.cpp) -- High Fidelity's serial interface to the ARM host board attached to the MPU-9150.
3. ["An efficient orientation filter for inertial/magnetic sensor arrays"](http://www.x-io.co.uk/res/doc/madgwick_internal_report.pdf) by Sebastian O.H. Madgwick
4. [Open source IMU and AHRS algorithms](http://www.x-io.co.uk/open-source-imu-and-ahrs-algorithms/)