# Assessment Service

The **Assessment Service** is a FastAPI-based microservice designed for managing psychological assessments, option configurations, and question templates, as well as scoring submissions and generating AI-powered advice.

## Database Migrations

Database schemas and seed data are managed using Alembic. Run the following commands to manage migrations:

```bash
# Apply schema and seed data
alembic upgrade head

# Tear down schema (deletes seeded rows and drops tables)
alembic downgrade base
```

## API Documentation

The interactive API documentation is available at `/swagger-ui` when the service is running. Below is a summary of the available endpoints:

### Health Check

*   **`GET /`**
    *   **Description:** Returns a simple service status message.
*   **`GET /health`**
    *   **Description:** Returns the health status of the assessment service.

### Management APIs

*   **`POST /api/v1/assessments/templates`**
    *   **Description:** Creates a new assessment template (e.g., Depression/Anxiety scale).
    *   **Request Body:** `TemplateCreate` (code, title, description).

*   **`POST /api/v1/assessments/option-groups`**
    *   **Description:** Configures grouped answer options (e.g., Likert scale) with labels, values, and scores.
    *   **Request Body:** `OptionGroupCreate` (group_code, group_name, description, options list).

*   **`POST /api/v1/assessments/questions/bulk`**
    *   **Description:** Bulk adds questions to an assessment template.
    *   **Request Body:** `QuestionBulkCreate` (template_id, group_id, questions list).

### User APIs

*   **`GET /api/v1/assessments/templates`**
    *   **Description:** Retrieves all active assessment templates.
    *   **Response:** List of templates.

*   **`GET /api/v1/assessments/templates/{template_id}/questions`**
    *   **Description:** Retrieves all questions and their corresponding options for a specific template.
    *   **Response:** List of questions with embedded options.

*   **`POST /api/v1/assessments/submit`**
    *   **Description:** Submits assessment answers. It calculates total/dimension scores and generates AI-driven psychological advice.
    *   **Headers:** `Authorization: Bearer <user_id>` (Required)
    *   **Request Body:** `AssessmentSubmit` (template_id, user_id, answers list).
    *   **Response:** Score calculation and AI evaluation feedback.
