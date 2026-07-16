# Coding Convention

This document defines the naming, style, and workflow conventions applied across the **MindCare** project.

---

## 1. Naming Rule

| Ingredient | Rule | Example |
|---|---|---|
| Folder | kebab-case | `user-profile` |
| File | kebab-case | `user-product.server.ts` |
| Class | PascalCase | `UserService` |
| Interface | Prefix `I` + PascalCase | `IUser` |
| Variable | camelCase | `userName` |
| Constant | UPPER_CASE_SNAKE_CASE | `MAX_REQUEST` |
| Enum | PascalCase | `OrderStatus` |
| Enum key | UPPER_CASE_SNAKE_CASE | `PHONE_NUMBER` |
| Enum value | kebab-case | `phone-number` |
| Function | camelCase | `updateOrder` |

---

## 2. Coding Style

* **Indent**: 2 spaces
* **Line length**: max 120 characters
* **Quotes**: use single quotes `' '`
* **Semicolon**: required (`;`)
* **Bracket spacing**: have a space after `{` and before `}`

---

## 3. Rules for Writing JS Code

* Do not use `any`, except when necessary.
* Always handle errors (use `try...catch`, etc.).
* Validate input.
* No hardcoded values.
* Do not use `console.log`.
* Use `class`, `interface`, or `type` to describe data.

---

## 4. Git Convention

### 4.1. Branch Name

Format: `<type feature>/<Task ID>`

Example: `features/TRE-111-BE`

### 4.2. Commit Message

Format:

```text
Parent ID: <parent name> (<commit number>)
Task Id: <Task ID>
Type: <type feature>
Description:
<description>
```

Example:

```text
Parent ID: TRE-282-BE (2)
Task Id: TRE-282-BE
Type: Development
Description:
- Remove code not use
```

> **Note:** `Type` includes: `features`, `bugs`.

---

## 5. Migration Rules

### 5.1. Adding a required field to a table that already has data

When a new field must be **required (`NOT NULL`)** on a table that already contains data:

1. Create a migration to add the field, allowing it to be **nullable**.
2. Create a migration to **backfill** data for the field on existing rows.
3. Create a migration to alter the field to **`NOT NULL`**.
