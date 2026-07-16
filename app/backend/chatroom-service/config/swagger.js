import swaggerJsdoc from "swagger-jsdoc";

const options = {
  definition: {
    openapi: "3.0.0",
    info: {
      title: "Chatroom Service API",
      version: "1.0.0",
      description:
        "REST surface of the MindCare chatroom-service (rooms listing, health). " +
        "Real-time chat, presence, direct messages, reactions and WebRTC call " +
        "signaling are delivered over Socket.IO, not REST — see README.md in " +
        "this service for the full Socket.IO event contract.",
    },
    servers: [
      { url: "/api/v1/chatroom", description: "Direct or via API gateway" },
    ],
  },
  apis: ["./routes/*.routes.js"],
};

export const swaggerSpec = swaggerJsdoc(options);
