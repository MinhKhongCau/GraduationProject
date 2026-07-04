package com.mindcare.auth.application.port.out;

public interface EmailSenderPort {
    void sendEmail(String to, String subject, String body);
}
