/** No chat backend exists yet (Chat Service is planned, not in app/backend). UI-only. */
export interface MockContact {
  id: string;
  name: string;
  avatar: string;
  lastMessage: string;
  time: string;
  unread: boolean;
  online: boolean;
}

export interface MockChatMessage {
  id: string;
  sender: "patient" | "expert";
  text: string;
  time: string;
}

export const CONTACTS_MOCK: MockContact[] = [
  {
    id: "1",
    name: "MSc. Nguyen An Binh",
    avatar: "https://i.pravatar.cc/150?u=binh",
    lastMessage: "Hello, I have received your information...",
    time: "10:30",
    unread: true,
    online: true,
  },
  {
    id: "2",
    name: "Dr. Tran Minh Tam",
    avatar: "https://i.pravatar.cc/150?u=tam",
    lastMessage: "Don't forget our regular consultation schedule...",
    time: "08:15",
    unread: false,
    online: false,
  },
  {
    id: "3",
    name: "MSc. Le Thi Hanh",
    avatar: "https://i.pravatar.cc/150?u=hanh",
    lastMessage: "Thank you for sharing your story with me.",
    time: "Yesterday",
    unread: false,
    online: true,
  },
  {
    id: "4",
    name: "PhD. Pham Quang Vinh",
    avatar: "https://i.pravatar.cc/150?u=vinh",
    lastMessage: "Please try the deep breathing exercises I...",
    time: "Mon",
    unread: false,
    online: false,
  },
  {
    id: "5",
    name: "Dr. Phan Hoang Long",
    avatar: "https://i.pravatar.cc/150?u=long",
    lastMessage: "Your new prescription has been sent to your...",
    time: "Sat",
    unread: false,
    online: true,
  },
];

export const CHAT_HISTORY_MOCK: MockChatMessage[] = [
  {
    id: "1",
    sender: "patient",
    text: "Hello Doctor, I just completed the psychological assessment and feel a bit anxious about the results.",
    time: "10:20",
  },
  {
    id: "2",
    sender: "expert",
    text: "Hello, I have received your information. Please don't worry too much, the assessment results are just the first step for us to better understand your current condition.",
    time: "10:22",
  },
  {
    id: "3",
    sender: "expert",
    text: "Your anxiety level is at a Moderate level. Do you frequently have trouble sleeping?",
    time: "10:23",
  },
  {
    id: "4",
    sender: "patient",
    text: "Yes, lately I've been tossing and turning and overthinking about work before falling asleep.",
    time: "10:25",
  },
  {
    id: "5",
    sender: "expert",
    text: "Your assessment results show a slightly high anxiety level. I recommend trying some relaxation techniques before bed. We can start with the 4-7-8 breathing exercise.",
    time: "10:30",
  },
];
