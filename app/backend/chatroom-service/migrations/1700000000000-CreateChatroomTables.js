export class CreateChatroomTables1700000000000 {
  name = "CreateChatroomTables1700000000000";

  async up(queryRunner) {
    // 1. Create rooms table
    await queryRunner.query(`
      CREATE TABLE IF NOT EXISTS rooms (
        id VARCHAR(50) PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        icon VARCHAR(10),
        users INT DEFAULT 0
      )
    `);

    // 2. Create messages table
    await queryRunner.query(`
      CREATE TABLE IF NOT EXISTS messages (
        id VARCHAR(100) PRIMARY KEY,
        room_id VARCHAR(50),
        dm_id VARCHAR(100),
        type VARCHAR(20) NOT NULL,
        text TEXT,
        audio TEXT,
        duration INT DEFAULT 0,
        mime_type VARCHAR(50),
        user_name VARCHAR(100),
        from_id VARCHAR(100) NOT NULL,
        to_id VARCHAR(100),
        created_at BIGINT NOT NULL
      )
    `);

    // 3. Create indices
    await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages (room_id, created_at DESC)`);
    await queryRunner.query(`CREATE INDEX IF NOT EXISTS idx_messages_dm_created ON messages (dm_id, created_at DESC)`);
  }

  async down(queryRunner) {
    await queryRunner.query(`DROP INDEX IF EXISTS idx_messages_dm_created`);
    await queryRunner.query(`DROP INDEX IF EXISTS idx_messages_room_created`);
    await queryRunner.query(`DROP TABLE IF EXISTS messages`);
    await queryRunner.query(`DROP TABLE IF EXISTS rooms`);
  }
}
