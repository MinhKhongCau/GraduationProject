package com.mindcare.auth.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.context.annotation.Configuration;

import java.util.HashMap;
import java.util.Map;

/**
 * Bind toàn bộ block `internal-clients:` trong application.yaml vào một Map.
 *
 * <pre>
 * # application.yaml
 * internal-clients:
 *   booking-service: "$2a$10$..."
 *   payment-service: "$2a$10$..."
 * </pre>
 *
 * Spring Boot tự động bind khi dùng @ConfigurationProperties(prefix = "internal-clients").
 */
@Configuration
@ConfigurationProperties(prefix = "internal-clients")
public class InternalClientsConfig {

    /**
     * Map: clientId → bcrypt-hashed-secret
     * Key phải match tên service trong application.yaml (vd: "booking-service")
     */
    private Map<String, String> clients = new HashMap<>();

    // Spring Boot cần getter/setter để bind
    public Map<String, String> getClients() {
        return clients;
    }

    public void setClients(Map<String, String> clients) {
        this.clients = clients;
    }

    /**
     * Spring Boot khi dùng prefix = "internal-clients" sẽ bind như sau:
     *   internal-clients.booking-service → clients["booking-service"]
     * Nhưng do prefix đã là "internal-clients", các key con sẽ bind trực tiếp
     * vào đây thông qua relaxed binding.
     *
     * Alternative: nếu yaml là:
     *   internal-clients:
     *     booking-service: "hash"
     * thì ta cần bind Map trực tiếp, không wrap thêm.
     */

    /**
     * Convenience method: lấy hash của một client_id.
     *
     * @param clientId vd: "booking-service"
     * @return bcrypt hash, hoặc null nếu không tồn tại
     */
    public String getHashedSecret(String clientId) {
        return clients.get(clientId);
    }
}
