class PuppetApplication
  
  constructor: ->
    @renderPane = $('#primary')
    @head = new PuppetHead(@renderPane)
    @lastQuaternion = [0, 0, 0, 1]
    @enabled = false

    @socket = new MotionSocket('ws://' + document.location.host + '/motion')
    @socket.bind 'motion', @updateQuaternion.bind(@)

    $("#reset-ahrs-btn").click =>
      @socket.resetAHRS()

    @socket.setPingCallback (ms)=>
      $('#latency-indicator').html("Ping #{ms}ms")

  updateQuaternion: (e)->
    @lastQuaternion = e.Data.quaternion

  start: ->
    @enabled = true
    @socket.subscribeMotion()
    @render()

  stop: ->
    @socket.unsubscribeMotion()
    @enabled = false

  render: ->
    if @enabled
      requestAnimationFrame(@render.bind(@))
      @head.updateQuaternion(@lastQuaternion...)
      @head.render()


class MotionSocket

  constructor: (hostString)->
    # Instantiate websocket.
    @ws = if window['MozWebSocket'] then new MozWebSocket(hostString) else new WebSocket(hostString)
    $(@ws).bind('message', @onMessage.bind(@))
    $(window).unload @close.bind(@)
    @handlers = {}
    # Setup ping interval timer.
    @pingLatency = 0
    
    @pingInterval = setInterval =>
      @send('ping', t: new Date().getTime())
    , 520

    @bind 'pong', (e)=>
      tStart = parseInt(e.Data.t)
      tEnd = new Date().getTime()
      @pingLatency = tEnd - tStart
      @pingIntervalUpdateCallback?(@pingLatency)
      console.log("Latency (ms):", @pingLatency)

  bind: (evt, f)->
    @handlers[evt] = f

  onMessage: (e)->
    m = JSON.parse(e.originalEvent.data)
    h = @handlers[m.Type]
    h?.call(this, m)

  send: (type, data)->
    m = {'Type': type}
    m['Data'] = data if data
    return @ws.send(JSON.stringify(m))

  close: ->
    @unsubscribeMotion()
    clearInterval(@pingInterval)
    @ws.close()

  resetAHRS: ->
    @send('reset_ahrs')

  setFPS: (fps)->
    @send('set_fps', fps)

  subscribeMotion: (fps)->
    @send('subscribe_motion')

  unsubscribeMotion: ->
    @send('unsubscribe_motion')

  setPingCallback: (f)->
    @pingIntervalUpdateCallback = f


class PuppetHead

  constructor: (@target)->
    @camera = new THREE.PerspectiveCamera( 75, window.innerWidth / window.innerHeight, 1, 10000 )
    @camera.position.z = 1000

    @scene = new THREE.Scene()

    @geometry = new THREE.SphereGeometry( 200, 30, 30 )
    @material = new THREE.MeshBasicMaterial( { color: 0xff0000, wireframe: true } )

    @mesh = new THREE.Mesh( @geometry, @material )
    @mesh.useQuaternion = true
    @scene.add( @mesh )

    @renderer = new THREE.CanvasRenderer()
    @renderer.setSize( @target.innerWidth(), @target.innerHeight() )

    @target.append(@renderer.domElement)

  updateQuaternion: (x, y, z, w)->
    @mesh.quaternion.set(x, y, z, w)

  render: ->
    @renderer.render(@scene, @camera)


# Make the PuppetApplication class globally available.
window.PuppetApplication = PuppetApplication