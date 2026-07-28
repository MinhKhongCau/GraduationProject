import { EntitySchema } from "typeorm";

export const MessageEntity = new EntitySchema({
  name: "Message",
  tableName: "messages",
  columns: {
    id: {
      primary: true,
      type: "varchar",
      length: 100,
    },
    room_id: {
      type: "varchar",
      length: 50,
      nullable: true,
    },
    dm_id: {
      type: "varchar",
      length: 100,
      nullable: true,
    },
    type: {
      type: "varchar",
      length: 20,
      nullable: false,
    },
    text: {
      type: "text",
      nullable: true,
    },
    audio: {
      type: "text",
      nullable: true,
    },
    duration: {
      type: "int",
      default: 0,
    },
    mime_type: {
      type: "varchar",
      length: 50,
      nullable: true,
    },
    user_name: {
      type: "varchar",
      length: 100,
      nullable: true,
    },
    from_id: {
      type: "varchar",
      length: 100,
      nullable: false,
    },
    to_id: {
      type: "varchar",
      length: 100,
      nullable: true,
    },
    created_at: {
      type: "bigint",
      nullable: false,
    },
  },
  indices: [
    {
      name: "idx_messages_room_created",
      columns: ["room_id", "created_at"],
    },
    {
      name: "idx_messages_dm_created",
      columns: ["dm_id", "created_at"],
    },
  ],
});
