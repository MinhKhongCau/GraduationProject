# this is file for read html and extract data
# get html by url using request library
# then extract data from html
# then save data to database
# then return data

import requests
from bs4 import BeautifulSoup
import fitz
import re
import uuid

class Accessment:
    def __init__(self, template_id, template_title, template_code, template_description):
        self.template_id = template_id
        self.template_title = template_title
        self.template_code = template_code
        self.template_description = template_description

    def __repr__(self):
        return f"Accessment(template_id={self.template_id}, template_title={self.template_title}, template_code={self.template_code}, template_description={self.template_description})"


class Question:
    def __init__(self, question_id, template_id, question_order, question_content):
        self.question_id = question_id
        self.template_id = template_id
        self.question_order = question_order
        self.question_content = question_content

    def __repr__(self):
        return f"Question(question_id={self.question_id}, template_id={self.template_id}, question_order={self.question_order}, question_content={self.question_content})"

class Option:
    def __init__(self, option_id, question_id, option_content, score_value):
        self.option_id = option_id
        self.question_id = question_id
        self.option_content = option_content
        self.score_value = score_value

    def __repr__(self):
        return f"Option(option_id={self.option_id}, question_id={self.question_id}, option_content={self.option_content}, score_value={self.score_value})"

class AccessmentParser:
    def __init__(self):
        self.access_template = None
        self.question_list = []
        self.option_list = []

    def read_data_from_url(self, url):
        # get html by url using request library
        response = requests.get(url)
        # then extract data from html
        soup = BeautifulSoup(response.text, 'html.parser')
        # then return data
        return soup
    

    def read_data_from_pdf(self, pdf_path):
        # read data from pdf
        doc = fitz.open(pdf_path)
        pages = [] 
        for page in doc:
            pages.append(page.get_text())
        # then return data
        return "\n".join(pages)

    def clean_pdf_content(self, text):
        # remove option number
        text = re.sub(r"\n\s*[0-3]\s*", " ", text)
        # remove newline
        text = re.sub(r"(\s+[0-3]){1,4}$", " ", text)
        # remove footer
        text = re.split(f"FOR OFFICE CODING", text, flags=re.I)[0]
        # then return data
        return text.strip()

    
    def parse_phq9_to_database_format(self, path="patient-health-questionnaire.pdf"):
        # read data from pdf
        data = self.read_data_from_pdf(path)
        # parse data to database format
        self.access_template = Accessment(
            str(uuid.uuid4()), 
            "Patient Health Questionnaire", 
            "PHQ-9", 
            "Patient Health Questionnaire"
        )
        # then return data
        options = []
        pattern = re.compile(r"(\d+)\.\s+(.*?)(?=\n\d+\.|\Z)", re.S)

        matches = pattern.findall(data)
        
        for order, content in matches:
            content = self.clean_pdf_content(content)
            question = Question(
                str(uuid.uuid4()),
                self.access_template.template_id,
                order,
                content.strip()
            )
            self.option_list = [
                Option(
                    str(uuid.uuid4()),
                    question.question_id,
                    option,
                    score
                ) for option, score in [
                    ("Not at all", 0),
                    ("Several days", 1),
                    ("More than half the days", 2),
                    ("Nearly every day", 3),
                ]
            ]
            self.question_list.append(question) 

        return {
            "access_template": self.access_template,
            "question_list": self.question_list,
            "option_list": self.option_list
        }

    def save_data_to_csv(data):
        # then save data to csv
        # then return data
        pass


    def save_data_to_database(url):
        # get html by url using request library
        response = requests.get(url)
        # then extract data from html
        soup = BeautifulSoup(response.text, 'html.parser')
        # then save data to database
        # then return data
        return soup


if __name__ == '__main__':
    # url = 'https://psychology-tools.com/test/obsessive-compulsive-inventory-revised'
    # parser = AccessmentParser("Obsessive Compulsive Inventory", "OCI-R")
    # data = AccessmentParser.read_data_from_url(url)
    path = "patient-health-questionnaire.pdf"
    access_parser = AccessmentParser()
    data = access_parser.parse_phq9_to_database_format(path)

    print(access_parser.access_template)
    print(access_parser.question_list)
    print(access_parser.option_list)
