import { EntitySchema } from "typeorm";

export const RoomEntity = new EntitySchema({
  name: "Room",
  tableName: "rooms",
  columns: {
    id: {
      primary: true,
      type: "varchar",
      length: 50,
    },
    name: {
      type: "varchar",
      length: 100,
      nullable: false,
    },
    icon: {
      type: "varchar",
      length: 10,
      nullable: true,
    },
    users: {
      type: "int",
      default: 0,
    },
  },
});
