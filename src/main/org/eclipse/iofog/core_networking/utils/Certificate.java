package main.org.eclipse.iofog.core_networking.utils;

import java.io.FileInputStream;
import java.security.cert.CertificateFactory;
import java.security.cert.X509Certificate;

/**
 * class to create @{@link java.security.cert.X509Certificate} from intermediate certificate file
 * <p>
 * Created by saeid on 4/8/16.
 */
public class Certificate {
    private X509Certificate certificate;

    public Certificate(String certFile) {
        try {
            FileInputStream is = new FileInputStream(certFile);
            CertificateFactory certificateFactory = CertificateFactory.getInstance("X.509");
            this.certificate = ((X509Certificate) certificateFactory.generateCertificate(is));
        } catch (Exception e) {
            this.certificate = null;
        }
    }

    public synchronized X509Certificate getCertificate() {
        return certificate;
    }
}
