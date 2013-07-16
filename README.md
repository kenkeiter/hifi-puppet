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

1. Ensure you have Go installed on your machine. You may either [download a pre-built binary](https://code.google.com/p/go/downloads/list) for your platform, or (if you're on OS X) install it via [Homebrew](http://mxcl.github.io/homebrew/) using the following command:

        $ brew install go

2. Checkout this repository using Git:

        $ git clone https://git.github.com/kenkeiter/hifi-puppet

3. Build the executable for your platform. The makefile will automatically download and install any dependencies -- but it may take a second. Literally.

        $ cd hifi-puppet/
        $ make && make install

4. Start the server:

        $ hifi-puppet

5. You're done! Open a webkit-compatible browser, and direct it to: [http://localhost:8192/](http://localhost:8192/)

## Configuration

### Command-Line Options

* `--host <host>` -- set the host and port upon which the interface will be served. This typically defaults to `0.0.0.0:8192`, and should follow a similar format.
* `--asset_path <path>` -- specifies a specific absolute path from which interface assets will be served. This currently defaults to `/var/lib/hifi-puppet`, which is where assets are installed by default.

### Connecting from Outside Your LAN

In order to allow users to connect to Puppet from outside your LAN, you'll need to do the following:

1. Disable the firewall on your computer, or be sure to allow access to port 8192 via TCP.
2. Enable port-forwarding on your router. You'll need to forward port 8192 from your WAN to your local machine.
3. Determine your public IP address (you can [find out from Google](https://www.google.com/search?q=whats+my+ip&oq=whats+my+ip)), and direct remote clients to `http://<your public IP>:8192/`

## References

1. [MPU9150Lib](https://github.com/Pansenti/MPU9150Lib/blob/master/libraries/MPU9150Lib/MPU9150Lib.cpp) -- A common library for the MPU-9150
2. [SerialInterface.cpp](https://github.com/worklist/hifi/blob/master/interface/src/SerialInterface.cpp) -- High Fidelity's serial interface for the Maple host board attached to the MPU-9150.