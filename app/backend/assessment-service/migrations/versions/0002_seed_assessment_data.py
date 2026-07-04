"""seed assessment templates data

Revision ID: 0002_seed_assessment_data
Revises: 0001_create_assessment_schema
Create Date: 2026-07-04

"""
import uuid

import sqlalchemy as sa
from alembic import op
from sqlalchemy.dialects.postgresql import UUID

# revision identifiers, used by Alembic.
revision = "0002_seed_assessment_data"
down_revision = "0001_create_assessment_schema"
branch_labels = None
depends_on = None

assess_templates = sa.table(
    "assess_templates",
    sa.column("template_id", UUID(as_uuid=True)),
    sa.column("code", sa.String),
    sa.column("title", sa.String),
    sa.column("description", sa.Text),
    sa.column("is_active", sa.Boolean),
)

assess_option_groups = sa.table(
    "assess_option_groups",
    sa.column("group_id", UUID(as_uuid=True)),
    sa.column("group_code", sa.String),
    sa.column("group_name", sa.String),
    sa.column("description", sa.Text),
)

assess_options = sa.table(
    "assess_options",
    sa.column("option_id", UUID(as_uuid=True)),
    sa.column("group_id", UUID(as_uuid=True)),
    sa.column("label", sa.String),
    sa.column("value", sa.String),
    sa.column("score_value", sa.Integer),
    sa.column("order_index", sa.Integer),
)

assess_questions = sa.table(
    "assess_questions",
    sa.column("question_id", UUID(as_uuid=True)),
    sa.column("template_id", UUID(as_uuid=True)),
    sa.column("group_id", UUID(as_uuid=True)),
    sa.column("content", sa.Text),
    sa.column("question_order", sa.Integer),
    sa.column("is_required", sa.Boolean),
)

# Group/template codes touched by this migration, used to clean up on
# downgrade without relying on the hardcoded ids assigned at upgrade time.
_TEMPLATE_CODES = (
    "PHQ_9",
    "GAD_7",
    "MDQ",
    "OCI_R",
    "ASRS_V1_1",
    "EAT_26",
    "PCL_5",
)
_GROUP_CODES = (
    "PHQ9_FREQ",
    "GAD_FREQ_4",
    "YES_NO",
    "SEVERITY_4",
    "LIKERT_0_4",
    "ASRS_FREQ_5",
    "EAT26_NORMAL",
    "EAT26_REVERSE",
    "PCL5_SCALE",
)


def _seed_phq9() -> None:
    template_id = uuid.uuid4()
    group_id = uuid.uuid4()

    op.bulk_insert(
        assess_templates,
        [
            {
                "template_id": template_id,
                "code": "PHQ_9",
                "title": "PHQ-9",
                "description": (
                    "Bảng câu hỏi sàng lọc và đánh giá mức độ trầm cảm "
                    "trong 2 tuần gần đây"
                ),
                "is_active": True,
            }
        ],
    )

    op.bulk_insert(
        assess_option_groups,
        [
            {
                "group_id": group_id,
                "group_code": "PHQ9_FREQ",
                "group_name": "Tần suất PHQ-9",
                "description": (
                    "Mức độ xuất hiện triệu chứng trong 2 tuần qua "
                    "(0–3 điểm)"
                ),
            }
        ],
    )

    options = [
        ("Không ngày nào", "0", 0),
        ("Vài ngày", "1", 1),
        ("Hơn một nửa số ngày", "2", 2),
        ("Gần như mọi ngày", "3", 3),
    ]
    op.bulk_insert(
        assess_options,
        [
            {
                "option_id": uuid.uuid4(),
                "group_id": group_id,
                "label": label,
                "value": value,
                "score_value": score,
                "order_index": i + 1,
            }
            for i, (label, value, score) in enumerate(options)
        ],
    )

    questions = [
        "Ít hứng thú hoặc không có niềm vui trong việc thực hiện các "
        "hoạt động hàng ngày.",
        "Cảm thấy buồn bã, chán nản, hoặc tuyệt vọng.",
        "Khó ngủ, ngủ chập chờn hoặc ngủ quá nhiều.",
        "Cảm thấy mệt mỏi hoặc thiếu năng lượng.",
        "Ăn không ngon miệng hoặc ăn quá nhiều.",
        "Cảm thấy tồi tệ về bản thân – hoặc thấy mình là người thất "
        "bại, làm cho bản thân hoặc gia đình thất vọng.",
        "Khó tập trung vào việc gì đó, chẳng hạn như khi đọc báo hoặc "
        "xem tivi.",
        "Di chuyển hoặc nói năng chậm chạp đến mức người khác có thể "
        "nhận thấy. Hoặc ngược lại, bồn chồn, không thể ngồi yên một "
        "chỗ.",
        "Có ý nghĩ rằng mình nên chết đi hoặc muốn làm bản thân bị "
        "thương theo một cách nào đó.",
    ]
    op.bulk_insert(
        assess_questions,
        [
            {
                "question_id": uuid.uuid4(),
                "template_id": template_id,
                "group_id": group_id,
                "content": content,
                "question_order": i + 1,
                "is_required": True,
            }
            for i, content in enumerate(questions)
        ],
    )


def _seed_eat26() -> None:
    template_id = uuid.uuid4()
    normal_group_id = uuid.uuid4()
    reverse_group_id = uuid.uuid4()

    op.bulk_insert(
        assess_templates,
        [
            {
                "template_id": template_id,
                "code": "EAT_26",
                "title": "EAT-26",
                "description": (
                    "Thang đo thái độ ăn uống (Eating Attitudes Test "
                    "– 26 câu)"
                ),
                "is_active": True,
            }
        ],
    )

    op.bulk_insert(
        assess_option_groups,
        [
            {
                "group_id": normal_group_id,
                "group_code": "EAT26_NORMAL",
                "group_name": "Mức độ EAT-26 (chuẩn)",
                "description": "Chấm điểm cho câu 1–25",
            },
            {
                "group_id": reverse_group_id,
                "group_code": "EAT26_REVERSE",
                "group_name": "Mức độ EAT-26 (đảo điểm)",
                "description": "Chấm điểm cho câu 26",
            },
        ],
    )

    normal_options = [
        ("Luôn luôn", "ALWAYS", 3),
        ("Thường xuyên", "OFTEN", 2),
        ("Khá thường xuyên", "SOMETIMES", 1),
        ("Đôi khi", "RARE", 0),
        ("Hiếm khi", "SELDOM", 0),
        ("Không bao giờ", "NEVER", 0),
    ]
    reverse_options = [
        ("Không bao giờ", "NEVER", 3),
        ("Hiếm khi", "SELDOM", 2),
        ("Đôi khi", "RARE", 1),
        ("Khá thường xuyên", "SOMETIMES", 0),
        ("Thường xuyên", "OFTEN", 0),
        ("Luôn luôn", "ALWAYS", 0),
    ]
    op.bulk_insert(
        assess_options,
        [
            {
                "option_id": uuid.uuid4(),
                "group_id": normal_group_id,
                "label": label,
                "value": value,
                "score_value": score,
                "order_index": i + 1,
            }
            for i, (label, value, score) in enumerate(normal_options)
        ]
        + [
            {
                "option_id": uuid.uuid4(),
                "group_id": reverse_group_id,
                "label": label,
                "value": value,
                "score_value": score,
                "order_index": i + 1,
            }
            for i, (label, value, score) in enumerate(reverse_options)
        ],
    )

    normal_questions = [
        "Tôi cực kỳ lo sợ về việc bị thừa cân.",
        "Tôi tránh ăn ngay cả khi tôi đang thấy đói.",
        "Tôi thấy mình luôn bận tâm quá mức đến thức ăn.",
        "Tôi có những lúc ăn uống vô độ và cảm thấy mình không thể "
        "dừng lại được.",
        "Tôi thường cắt nhỏ thức ăn thành những miếng rất bé.",
        "Tôi rất chú ý đến hàm lượng calo trong những thực phẩm mình "
        "ăn.",
        "Tôi đặc biệt tránh những thực phẩm có hàm lượng carbohydrate "
        "cao.",
        "Tôi cảm thấy người khác thích tôi ăn nhiều hơn mức hiện tại.",
        "Tôi bị nôn (mửa) sau khi ăn xong.",
        "Tôi cảm thấy cực kỳ tội lỗi sau khi ăn.",
        "Tôi luôn bị thôi thúc bởi mong muốn được gầy hơn nữa.",
        "Tôi nghĩ về việc đốt cháy calo mỗi khi tập thể dục.",
        "Những người khác nghĩ rằng tôi quá gầy.",
        "Tôi luôn bận tâm với ý nghĩ về việc cơ thể mình có mỡ.",
        "Tôi mất nhiều thời gian để ăn xong một bữa hơn những người "
        "khác.",
        "Tôi tránh các loại thực phẩm có chứa đường.",
        "Tôi thường xuyên ăn các loại thực phẩm ăn kiêng.",
        "Tôi cảm thấy thức ăn đang kiểm soát cuộc sống của mình.",
        "Tôi thể hiện sự tự kiểm soát rất cao đối với việc ăn uống.",
        "Tôi cảm thấy những người khác đang ép tôi phải ăn.",
        "Tôi dành quá nhiều thời gian và tâm trí cho thức ăn.",
        "Tôi cảm thấy không thoải mái sau khi ăn đồ ngọt.",
        "Tôi thường xuyên thực hiện các hành vi ăn kiêng.",
        "Tôi thích cảm giác khi bụng mình đang trống rỗng.",
        "Tôi có thôi thúc muốn nôn mửa sau các bữa ăn.",
    ]
    op.bulk_insert(
        assess_questions,
        [
            {
                "question_id": uuid.uuid4(),
                "template_id": template_id,
                "group_id": normal_group_id,
                "content": content,
                "question_order": i + 1,
                "is_required": True,
            }
            for i, content in enumerate(normal_questions)
        ]
        + [
            {
                "question_id": uuid.uuid4(),
                "template_id": template_id,
                "group_id": reverse_group_id,
                "content": "Tôi thích thử những loại thức ăn mới lạ "
                "và giàu dinh dưỡng.",
                "question_order": 26,
                "is_required": True,
            }
        ],
    )


def _seed_pcl5() -> None:
    template_id = uuid.uuid4()
    group_id = uuid.uuid4()

    op.bulk_insert(
        assess_templates,
        [
            {
                "template_id": template_id,
                "code": "PCL_5",
                "title": "PCL-5",
                "description": (
                    "Bảng câu hỏi PTSD Checklist for DSM-5 (20 câu)"
                ),
                "is_active": True,
            }
        ],
    )

    op.bulk_insert(
        assess_option_groups,
        [
            {
                "group_id": group_id,
                "group_code": "PCL5_SCALE",
                "group_name": "Mức độ PCL-5",
                "description": (
                    "Mức độ bị làm phiền trong tháng vừa qua "
                    "(0–4 điểm)"
                ),
            }
        ],
    )

    options = [
        ("Không chút nào", "0", 0),
        ("Một ít", "1", 1),
        ("Vừa phải", "2", 2),
        ("Nhiều", "3", 3),
        ("Rất nhiều", "4", 4),
    ]
    op.bulk_insert(
        assess_options,
        [
            {
                "option_id": uuid.uuid4(),
                "group_id": group_id,
                "label": label,
                "value": value,
                "score_value": score,
                "order_index": i + 1,
            }
            for i, (label, value, score) in enumerate(options)
        ],
    )

    questions = [
        "Những ký ức lặp đi lặp lại, không mong muốn và gây đau khổ "
        "về trải nghiệm sang chấn?",
        "Những giấc mơ lặp đi lặp lại và gây đau khổ về trải nghiệm "
        "sang chấn?",
        "Bỗng nhiên cảm thấy hoặc hành động như thể trải nghiệm sang "
        "chấn đang thực sự diễn ra một lần nữa?",
        "Cảm thấy rất buồn khổ khi có thứ gì đó gợi nhắc đến trải "
        "nghiệm sang chấn?",
        "Có các phản ứng cơ thể mạnh mẽ khi bị gợi nhắc về trải "
        "nghiệm sang chấn?",
        "Tránh suy nghĩ hoặc cảm xúc liên quan đến trải nghiệm sang "
        "chấn?",
        "Tránh những nhắc nhở bên ngoài về trải nghiệm sang chấn?",
        "Không thể nhớ lại những phần quan trọng của trải nghiệm sang "
        "chấn?",
        "Có những niềm tin tiêu cực quá mức và dai dẳng về bản thân, "
        "người khác hoặc thế giới?",
        "Tự đổ lỗi cho bản thân hoặc người khác về trải nghiệm sang "
        "chấn hoặc những hậu quả của nó?",
        "Thường xuyên có những cảm xúc tiêu cực mạnh mẽ như sợ hãi, "
        "tức giận hoặc tội lỗi?",
        "Giảm hứng thú rõ rệt đối với các hoạt động mà bạn từng yêu "
        "thích?",
        "Cảm thấy xa cách hoặc bị cô lập với những người khác?",
        "Không thể có những cảm xúc tích cực?",
        "Hành vi cáu gắt, bùng nổ cơn giận hoặc hung hăng?",
        "Quá thận trọng hoặc luôn trong trạng thái cảnh giác cao độ?",
        "Cảm thấy giật mình hoặc hoảng hốt một cách dễ dàng?",
        "Gặp khó khăn trong việc tập trung?",
        "Khó ngủ hoặc ngủ không yên giấc?",
        "Có các hành vi liều lĩnh hoặc tự gây hại cho bản thân?",
    ]
    op.bulk_insert(
        assess_questions,
        [
            {
                "question_id": uuid.uuid4(),
                "template_id": template_id,
                "group_id": group_id,
                "content": content,
                "question_order": i + 1,
                "is_required": True,
            }
            for i, content in enumerate(questions)
        ],
    )


def _seed_gad7() -> None:
    # Already generates valid ids server-side via gen_random_uuid(),
    # so this block is carried over unchanged from Assessment_data.sql.
    op.execute(
        """
        INSERT INTO Assess_Templates (
          template_id, code, title, description, is_active, created_at
        ) VALUES (
          gen_random_uuid(),
          'GAD_7',
          'GAD-7 – Thang đo lo âu lan tỏa',
          'GAD-7 dùng để sàng lọc và đánh giá mức độ rối loạn lo âu trong 2 tuần gần nhất.',
          true,
          now()
        );

        INSERT INTO Assess_Option_Groups (
          group_id, group_code, group_name, description
        ) VALUES (
          gen_random_uuid(),
          'GAD_FREQ_4',
          'Tần suất 4 mức (GAD-7)',
          'Không ngày nào → Gần như mọi ngày (0–3 điểm)'
        );

        INSERT INTO Assess_Options (
          option_id, group_id, label, value, score_value, order_index
        )
        SELECT gen_random_uuid(), g.group_id, o.label, o.value, o.score, o.ord
        FROM Assess_Option_Groups g,
        (
          VALUES
            ('Không ngày nào', '0', 0, 1),
            ('Vài ngày', '1', 1, 2),
            ('Hơn một nửa số ngày', '2', 2, 3),
            ('Gần như mọi ngày', '3', 3, 4)
        ) AS o(label, value, score, ord)
        WHERE g.group_code = 'GAD_FREQ_4';

        INSERT INTO Assess_Questions (
          question_id,
          template_id,
          group_id,
          content,
          question_order,
          is_required
        )
        SELECT
          gen_random_uuid(),
          t.template_id,
          g.group_id,
          q.content,
          q.ord,
          true
        FROM Assess_Templates t
        JOIN Assess_Option_Groups g
          ON g.group_code = 'GAD_FREQ_4'
        CROSS JOIN (
          VALUES
            (1, 'Cảm thấy căng thẳng, lo lắng hoặc dễ bị kích động.'),
            (2, 'Không thể ngừng lo lắng hoặc không kiểm soát được sự lo lắng.'),
            (3, 'Lo lắng quá nhiều về những việc khác nhau.'),
            (4, 'Gặp khó khăn trong việc thư giãn.'),
            (5, 'Bồn chồn đến mức khó có thể ngồi yên một chỗ.'),
            (6, 'Trở nên dễ cáu gắt hoặc bực bội.'),
            (7, 'Cảm thấy sợ hãi như thể có điều gì đó tồi tệ sắp xảy ra.')
        ) AS q(ord, content)
        WHERE t.code = 'GAD_7';
        """
    )


def _seed_mdq() -> None:
    op.execute(
        """
        INSERT INTO Assess_Templates (
          template_id, code, title, description, is_active, created_at
        ) VALUES (
          gen_random_uuid(),
          'MDQ',
          'Mood Disorder Questionnaire (MDQ)',
          'Bảng câu hỏi sàng lọc rối loạn lưỡng cực (Mood Disorder Questionnaire)',
          true,
          now()
        );

        INSERT INTO Assess_Option_Groups (
          group_id, group_code, group_name, description
        ) VALUES (
          gen_random_uuid(),
          'YES_NO',
          'Có / Không',
          'Câu trả lời dạng Có hoặc Không'
        );

        INSERT INTO Assess_Option_Groups (
          group_id, group_code, group_name, description
        ) VALUES (
          gen_random_uuid(),
          'SEVERITY_4',
          'Mức độ ảnh hưởng',
          'Không / Nhẹ / Vừa / Nghiêm trọng'
        );

        INSERT INTO Assess_Options (
          option_id, group_id, label, value, score_value, order_index
        )
        SELECT gen_random_uuid(), g.group_id, o.label, o.val, o.score, o.ord
        FROM Assess_Option_Groups g,
        (
          VALUES
            ('Không', 'NO', 0, 1),
            ('Có', 'YES', 1, 2)
        ) o(label, val, score, ord)
        WHERE g.group_code = 'YES_NO';

        INSERT INTO Assess_Options (
          option_id, group_id, label, value, score_value, order_index
        )
        SELECT gen_random_uuid(), g.group_id, o.label, o.val, o.score, o.ord
        FROM Assess_Option_Groups g,
        (
          VALUES
            ('Không', 'NONE', 0, 1),
            ('Nhẹ', 'MILD', 0, 2),
            ('Vừa', 'MODERATE', 1, 3),
            ('Nghiêm trọng', 'SEVERE', 1, 4)
        ) o(label, val, score, ord)
        WHERE g.group_code = 'SEVERITY_4';

        INSERT INTO Assess_Questions (
          question_id,
          template_id,
          group_id,
          content,
          question_order,
          is_required
        )
        SELECT
          gen_random_uuid(),
          t.template_id,
          g.group_id,
          q.content,
          q.ord,
          true
        FROM Assess_Templates t
        CROSS JOIN (
          VALUES
            (1, 'YES_NO', 'Bạn cảm thấy quá tốt hoặc hưng phấn đến mức người khác nghĩ bạn không bình thường, hoặc bạn quá hưng phấn đến mức gặp rắc rối?'),
            (2, 'YES_NO', 'Bạn dễ kích động đến mức quát mắng mọi người hoặc bắt đầu các cuộc tranh cãi, đánh nhau?'),
            (3, 'YES_NO', 'Bạn cảm thấy tự tin hơn nhiều so với bình thường?'),
            (4, 'YES_NO', 'Bạn ngủ ít hơn nhiều so với bình thường và thấy rằng mình không thực sự cần ngủ thêm?'),
            (5, 'YES_NO', 'Bạn nói nhiều hơn hoặc nói nhanh hơn nhiều so với bình thường?'),
            (6, 'YES_NO', 'Các suy nghĩ chạy dồn dập trong đầu hoặc bạn không thể làm chậm tâm trí mình lại?'),
            (7, 'YES_NO', 'Bạn dễ dàng bị xao nhãng đến mức khó tập trung hoặc duy trì công việc đang làm?'),
            (8, 'YES_NO', 'Bạn có nhiều năng lượng hơn nhiều so với bình thường?'),
            (9, 'YES_NO', 'Bạn hoạt động nhiều hơn hoặc làm nhiều việc hơn nhiều so với bình thường?'),
            (10, 'YES_NO', 'Bạn cởi mở hoặc thích giao du hơn nhiều so với bình thường?'),
            (11, 'YES_NO', 'Bạn quan tâm đến tình dục nhiều hơn nhiều so với bình thường?'),
            (12, 'YES_NO', 'Bạn làm những việc không bình thường hoặc mạo hiểm?'),
            (13, 'YES_NO', 'Việc tiêu xài tiền bạc gây rắc rối cho bạn hoặc gia đình bạn?'),
            (14, 'YES_NO', 'Các hiện tượng trên có bao giờ xảy ra trong cùng một khoảng thời gian không?'),
            (15, 'SEVERITY_4', 'Những vấn đề trên gây ra rắc rối ở mức độ nào cho bạn?'),
            (16, 'YES_NO', 'Có người thân cùng huyết thống bị rối loạn lưỡng cực không?'),
            (17, 'YES_NO', 'Bạn đã từng được chuyên gia y tế chẩn đoán rối loạn lưỡng cực chưa?')
        ) AS q(ord, group_code, content)
        JOIN Assess_Option_Groups g
          ON g.group_code = q.group_code
        WHERE t.code = 'MDQ';
        """
    )


def _seed_oci_r() -> None:
    op.execute(
        """
        INSERT INTO Assess_Templates (
          template_id, code, title, description, is_active, created_at
        ) VALUES (
          gen_random_uuid(),
          'OCI_R',
          'Obsessive-Compulsive Inventory – Revised (OCI-R)',
          'Thang đo đánh giá mức độ triệu chứng rối loạn ám ảnh cưỡng chế (OCD)',
          true,
          now()
        );

        INSERT INTO Assess_Option_Groups (
          group_id, group_code, group_name, description
        ) VALUES (
          gen_random_uuid(),
          'LIKERT_0_4',
          'Mức độ 0–4',
          '0: Không chút nào → 4: Rất nhiều'
        );

        INSERT INTO Assess_Options (
          option_id, group_id, label, value, score_value, order_index
        )
        SELECT gen_random_uuid(), g.group_id, o.label, o.val, o.score, o.ord
        FROM Assess_Option_Groups g,
        (
          VALUES
            ('Không chút nào', '0', 0, 1),
            ('Một chút', '1', 1, 2),
            ('Vừa phải', '2', 2, 3),
            ('Nhiều', '3', 3, 4),
            ('Rất nhiều', '4', 4, 5)
        ) o(label, val, score, ord)
        WHERE g.group_code = 'LIKERT_0_4';

        INSERT INTO Assess_Questions (
          question_id,
          template_id,
          group_id,
          content,
          question_order,
          is_required
        )
        SELECT
          gen_random_uuid(),
          t.template_id,
          g.group_id,
          q.content,
          q.ord,
          true
        FROM Assess_Templates t
        CROSS JOIN (
          VALUES
            (1, 'Tôi tích trữ rất nhiều thứ đến nỗi chúng làm cản trở không gian sống.'),
            (2, 'Tôi kiểm tra mọi thứ thường xuyên hơn mức cần thiết.'),
            (3, 'Tôi thấy khó chịu nếu các đồ vật không được sắp xếp đúng cách.'),
            (4, 'Tôi cảm thấy buộc phải đếm nhẩm trong khi đang làm việc gì đó.'),
            (5, 'Tôi thấy khó khăn khi chạm vào một đồ vật nếu biết người lạ hoặc ai đó đã chạm vào nó.'),
            (6, 'Tôi thấy khó kiểm soát được những suy nghĩ của chính mình.'),
            (7, 'Tôi thu thập những thứ mà tôi không thực sự cần đến.'),
            (8, 'Tôi lặp đi lặp lại việc kiểm tra cửa sổ, cửa ra vào, ngăn kéo, v.v.'),
            (9, 'Tôi bực bội nếu người khác thay đổi cách tôi đã sắp xếp đồ đạc.'),
            (10, 'Tôi cảm thấy mình phải lặp lại một số con số nhất định.'),
            (11, 'Đôi khi tôi phải tắm rửa hoặc làm sạch bản thân chỉ vì cảm thấy mình bị "bẩn".'),
            (12, 'Tôi bị phiền lòng bởi những suy nghĩ khó chịu xuất hiện trong đầu trái với ý muốn của mình.'),
            (13, 'Tôi tránh vứt đồ đạc đi vì sợ rằng sau này mình có thể cần đến chúng.'),
            (14, 'Tôi lặp đi lặp lại việc kiểm tra vòi gas, vòi nước và công tắc điện sau khi đã tắt chúng.'),
            (15, 'Tôi cần đồ vật phải được sắp xếp theo một cách cụ thể nào đó.'),
            (16, 'Tôi cảm thấy có những con số "tốt" và những con số "xấu".'),
            (17, 'Tôi rửa tay thường xuyên hơn và lâu hơn mức cần thiết.'),
            (18, 'Tôi thường xuyên có những suy nghĩ tục tĩu/xấu xa và gặp khó khăn trong việc xua tan chúng.')
        ) q(ord, content)
        JOIN Assess_Option_Groups g
          ON g.group_code = 'LIKERT_0_4'
        WHERE t.code = 'OCI_R';
        """
    )


def _seed_asrs() -> None:
    op.execute(
        """
        INSERT INTO Assess_Templates (
          template_id, code, title, description, is_active, created_at
        ) VALUES (
          gen_random_uuid(),
          'ASRS_V1_1',
          'Adult ADHD Self-Report Scale (ASRS v1.1)',
          'Thang đo sàng lọc ADHD ở người trưởng thành (ASRS v1.1)',
          true,
          now()
        );

        INSERT INTO Assess_Option_Groups (
          group_id, group_code, group_name, description
        ) VALUES (
          gen_random_uuid(),
          'ASRS_FREQ_5',
          'Tần suất 5 mức (ASRS)',
          'Không bao giờ → Rất thường xuyên'
        );

        INSERT INTO Assess_Options (
          option_id, group_id, label, value, score_value, order_index
        )
        SELECT gen_random_uuid(), g.group_id, o.label, o.val, o.score, o.ord
        FROM Assess_Option_Groups g,
        (
          VALUES
            ('Không bao giờ', 'NEVER', 0, 1),
            ('Hiếm khi', 'RARELY', 0, 2),
            ('Đôi khi', 'SOMETIMES', 1, 3),
            ('Thường xuyên', 'OFTEN', 2, 4),
            ('Rất thường xuyên', 'VERY_OFTEN', 3, 5)
        ) o(label, val, score, ord)
        WHERE g.group_code = 'ASRS_FREQ_5';

        INSERT INTO Assess_Questions (
          question_id,
          template_id,
          group_id,
          content,
          question_order,
          is_required
        )
        SELECT
          gen_random_uuid(),
          t.template_id,
          g.group_id,
          q.content,
          q.ord,
          true
        FROM Assess_Templates t
        CROSS JOIN (
          VALUES
            (1, 'Bạn có thường gặp khó khăn trong việc hoàn thành các chi tiết cuối cùng của một dự án, sau khi những phần thử thách nhất đã được thực hiện xong không?'),
            (2, 'Bạn có thường gặp khó khăn trong việc sắp xếp công việc theo thứ tự khi phải thực hiện một nhiệm vụ đòi hỏi tính tổ chức không?'),
            (3, 'Bạn có thường gặp vấn đề trong việc ghi nhớ các cuộc hẹn hoặc các nghĩa vụ phải thực hiện không?'),
            (4, 'Khi có một nhiệm vụ đòi hỏi phải suy nghĩ nhiều, bạn có thường né tránh hoặc trì hoãn việc bắt đầu không?'),
            (5, 'Bạn có thường xuyên máy động tay chân hoặc cựa quậy khi phải ngồi yên trong một thời gian dài không?'),
            (6, 'Bạn có thường cảm thấy quá năng động và buộc phải làm việc gì đó, giống như thể bạn đang bị thúc đẩy bởi một bộ máy không?'),
            (7, 'Bạn có thường mắc phải những lỗi bất cẩn khi phải làm việc trong một dự án nhàm chán hoặc khó khăn không?'),
            (8, 'Bạn có thường gặp khó khăn trong việc duy trì sự chú ý khi đang làm những công việc nhàm chán hoặc lặp đi lặp lại không?'),
            (9, 'Bạn có thường gặp khó khăn trong việc tập trung vào những gì người khác nói với bạn, ngay cả khi họ đang nói chuyện trực tiếp với bạn không?'),
            (10, 'Bạn có thường để thất lạc hoặc gặp khó khăn khi tìm đồ đạc ở nhà hoặc nơi làm việc không?'),
            (11, 'Bạn có thường xuyên bị xao nhãng bởi các hoạt động hoặc tiếng ồn xung quanh không?'),
            (12, 'Bạn có thường rời khỏi chỗ ngồi trong các cuộc họp hoặc các tình huống khác mà bạn được yêu cầu phải ngồi yên không?'),
            (13, 'Bạn có thường xuyên cảm thấy bồn chồn hoặc đứng ngồi không yên không?'),
            (14, 'Bạn có thường gặp khó khăn trong việc thả lỏng và thư giãn khi có thời gian rảnh cho bản thân không?'),
            (15, 'Bạn có thấy mình thường nói quá nhiều khi ở trong các tình huống giao tiếp xã hội không?'),
            (16, 'Khi đang trò chuyện, bạn có thường thấy mình kết thúc câu nói của người đang nói chuyện với bạn trước khi họ tự hoàn thành nó không?'),
            (17, 'Bạn có thường gặp khó khăn khi phải chờ đến lượt mình trong các tình huống cần phải xếp hàng không?'),
            (18, 'Bạn có thường ngắt lời người khác khi họ đang bận không?')
        ) q(ord, content)
        JOIN Assess_Option_Groups g
          ON g.group_code = 'ASRS_FREQ_5'
        WHERE t.code = 'ASRS_V1_1';
        """
    )


def upgrade() -> None:
    _seed_phq9()
    _seed_gad7()
    _seed_mdq()
    _seed_oci_r()
    _seed_asrs()
    _seed_eat26()
    _seed_pcl5()


def downgrade() -> None:
    template_codes = "'" + "','".join(_TEMPLATE_CODES) + "'"
    group_codes = "'" + "','".join(_GROUP_CODES) + "'"

    op.execute(
        f"""
        DELETE FROM assess_questions
        WHERE template_id IN (
            SELECT template_id FROM assess_templates
            WHERE code IN ({template_codes})
        );
        """
    )
    op.execute(
        f"""
        DELETE FROM assess_options
        WHERE group_id IN (
            SELECT group_id FROM assess_option_groups
            WHERE group_code IN ({group_codes})
        );
        """
    )
    op.execute(
        f"""
        DELETE FROM assess_option_groups
        WHERE group_code IN ({group_codes});
        """
    )
    op.execute(
        f"""
        DELETE FROM assess_templates
        WHERE code IN ({template_codes});
        """
    )
