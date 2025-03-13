import mqtt from 'mqtt'

const options: mqtt.IClientOptions = {
  protocolVersion: 5,
  username: "melon",
  password: "password2"
}

const client = mqtt.connect("mqtt://127.0.0.1:1883", options)

client.on("connect", function () {
  console.log("connection successful")

  client.subscribe("updates/test", function (err, granted) {
    if (err) {
      console.log("Failed to subscribe to topic: 'updates/test'")
    }
  })

  client.subscribe("melon/test", function (err, granted) {
    if (err) {
      console.log("Failed to subscribe to topic: 'melon/test'")
    }
  })

  client.publish('updates/test', "check for write access", function (err, packet) {
    if (err) {
      console.error('Failed to publish message:', err);
    } else {
      console.log('Message published with retain flag set to true', packet);
    }
  })
})

client.on("message", function (topic, message) {
  console.log(`Received message on topic ${topic}: ${message}`)
})

client.on("offline", function () {
  console.log("Client is offline")
})

client.on("reconnect", function () {
  console.log("Reconnecting to MQTT broker")
})

client.on("end", function () {
  console.log("Connection to MQTT broker ended")
})