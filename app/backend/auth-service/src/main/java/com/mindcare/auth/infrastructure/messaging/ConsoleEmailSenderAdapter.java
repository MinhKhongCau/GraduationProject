package com.mindcare.auth.infrastructure.messaging;
import com.mindcare.auth.application.port.out.EmailSenderPort;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class ConsoleEmailSenderAdapter implements EmailSenderPort {
    @Override
    public void sendEmail(String to, String subject, String body) {
        log.info("========== EMAIL SENDER MOCK ==========");
        log.info("To: {}", to);
        log.info("Subject: {}", subject);
        log.info("Body: {}", body);
        log.info("=======================================");
    }
}
