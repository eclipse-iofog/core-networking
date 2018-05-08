package org.eclipse.iofog.core_networking.main;

import org.eclipse.iofog.api.IOFabricClient;
import org.eclipse.iofog.api.listener.IOFabricAPIListener;
import io.netty.handler.ssl.SslContext;
import io.netty.handler.ssl.SslContextBuilder;
import io.netty.util.internal.StringUtil;
import org.eclipse.iofog.core_networking.comsat_client.ComSatClient;
import org.eclipse.iofog.core_networking.comsat_client.ComSatClientThreadFactory;
import org.eclipse.iofog.core_networking.utils.Certificate;
import org.eclipse.iofog.core_networking.utils.ContainerConfig;

import java.util.concurrent.ThreadFactory;
import java.util.logging.Logger;

/**
 * Created by saeid on 4/8/16.
 */
public class CoreNetworking {

    public static ContainerConfig config = null;
    public static String containerId = "";
    public static IOFabricClient ioFabricClient;

    private final Logger log = Logger.getLogger(CoreNetworking.class.getName());
    private ComSatClient[] connections;
    private Certificate cert;
    private boolean connecting;
    private IOFabricAPIListener listener;
    private SslContext sslCtx;
    private Object fetchConfigLock = new Object();

    public static void main(String[] args) throws Exception {
        CoreNetworking instance = new CoreNetworking();

        if (args.length > 0 && args[0].startsWith("--id=")) {
            CoreNetworking.containerId = args[0].substring(args[0].indexOf('=') + 1);
        } else {
            CoreNetworking.containerId = System.getenv("SELFNAME");
        }

        if (StringUtil.isNullOrEmpty(CoreNetworking.containerId)) {
            instance.log.warning("container id has not been set");
            instance.log.warning("use --id=XXXX parameter or set the id as SELFNAME=XXXX environment property");
            System.exit(1);
        }

        instance.init();
    }

    private void getCertificate() {
        cert = new Certificate("/jar-file/intermediate.crt");
        if (cert == null) {
            log.warning("error importing certificate.");
            System.exit(1);
        }
        try {
            sslCtx = SslContextBuilder
                    .forClient()
                    .trustManager(cert.getCertificate())
                    .build();
        } catch (Exception e) {
            log.warning(String.format("#%d : %s", cert, e.getMessage()));
        }

    }

    private void makeConnections() {
        ThreadFactory threadFactory = new ComSatClientThreadFactory();
        int connectionCount = CoreNetworking.config.getConnectionCount();
        connections = new ComSatClient[connectionCount];
        for (int counter = 0; counter < connectionCount; counter++) {
            while (isConnecting()) {
                try {
                    Thread.sleep(100);
                } catch (Exception e) {
                }
            }
            connecting();
            connections[counter] = new ComSatClient(sslCtx, this, counter);
            threadFactory.newThread(connections[counter]).start();
        }
    }

    private void closeAllConnections() {
        for (int counter = 0; counter < CoreNetworking.config.getConnectionCount(); counter++) {
            if (connections[counter] != null)
                connections[counter].close(true);
        }
    }

    private void start() {
        if (config.getMode().equals("private")) {
            try {
                ioFabricClient.openMessageWebSocket(listener);
            } catch (Exception e) {
                log.warning("unable to open message websocket");
                log.warning(e.getMessage());
                System.exit(1);
            }
        } else {
            while (config.getMode().equals("")) {
                try {
                    Thread.sleep(2000);
                } catch (Exception e) {
                }
            }
        }

        makeConnections();

        while (true) {
            try {
                for (int counter = 0; counter < CoreNetworking.config.getConnectionCount(); counter++)
                    connections[counter].beat();
                Thread.sleep(CoreNetworking.config.getHeartbeatFrequency());
            } catch (Exception e) {
            }
        }
    }

    private void init() {
        getCertificate();

        String ioFabricHost = System.getProperty("iofabric_host", "iofabric");
        int ioFabricPort = 54321;
        try {
            ioFabricPort = Integer.parseInt(System.getProperty("iofabric_port", "54321"));
        } catch (Exception e) {
        }

        ioFabricClient = new IOFabricClient(ioFabricHost, ioFabricPort, CoreNetworking.containerId);
        listener = new APIListenerImpl(this);

        fetchConfig();

        try {
            ioFabricClient.openControlWebSocket(listener);
        } catch (Exception e) {
            log.warning("unable to open control websocket");
            log.warning(e.getMessage());
            System.exit(1);
        }

        start();
    }

    public void connectingDone() {
        synchronized (CoreNetworking.class) {
            connecting = false;
        }
    }

    public boolean isConnecting() {
        synchronized (CoreNetworking.class) {
            return connecting;
        }
    }

    public void connecting() {
        synchronized (CoreNetworking.class) {
            connecting = true;
        }
    }

    private void fetchConfig() {
        config = null;
        try {
            while (config == null) {
                ioFabricClient.fetchContainerConfig(listener);
                synchronized (fetchConfigLock) {
                    fetchConfigLock.wait(1000);
                }
            }
        } catch (Exception e) {
            log.warning("unable to fetch config");
            log.warning(e.getMessage());
            System.exit(1);
        }

    }

    public void setConfig(ContainerConfig config) {
        CoreNetworking.config = config;
        synchronized (fetchConfigLock) {
            fetchConfigLock.notifyAll();
        }
    }

    public void updateConfig() {
        try {
            log.info("new config received");
            fetchConfig();
            closeAllConnections();
            ioFabricClient.openControlWebSocket(listener);
            start();
        } catch (Exception e) {
        }
    }

}
