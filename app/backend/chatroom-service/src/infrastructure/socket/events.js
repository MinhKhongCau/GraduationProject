export const EVENTS = {
    ROOM_JOIN: "room:join",
    ROOM_USERS: "room:users",
  
    CHAT_SEND: "chat:send",
    CHAT_MESSAGE: "chat:message",
    CHAT_HISTORY: "chat:history",
  
    WEBRTC_OFFER: "webrtc:offer",
    WEBRTC_ANSWER: "webrtc:answer",
    WEBRTC_ICE: "webrtc:ice",
  
    PRESENCE_SPEAKING: "presence:speaking",
    PRESENCE_SPEAKING_LIST: "presence:speaking:list",

    TYPING: "chat:typing",
    TYPING_STATUS: "chat:typing:status",

    // Meet room for a paid appointment (joined through the meeting link).
    MEET_JOIN: "meet:join",
    MEET_LEAVE: "meet:leave",
    MEET_SIGNAL: "meet:signal",
    MEET_PRESENCE: "meet:presence",
    MEET_END: "meet:end",
    MEET_PARTICIPANT_JOINED: "meet:participant-joined",
    MEET_PARTICIPANT_LEFT: "meet:participant-left",
    MEET_CAN_COMPLETE: "meet:can-complete",
    MEET_ENDED: "meet:ended",
  };