package org.eclipse.iofog.core.networking.local.client.private_client;

import org.eclipse.iofog.core.networking.local.client.LocalClient;
import org.eclipse.iofog.elements.IOMessage;
import io.netty.channel.Channel;
import org.eclipse.iofog.core.networking.main.CoreNetworking;
import org.eclipse.iofog.core.networking.utils.Constants;
import org.eclipse.iofog.core.networking.utils.MessageRepository;

import java.io.ByteArrayOutputStream;
import java.util.logging.Logger;

/**
 * communications in "private" mode
 * <p>
 * Created by saeid on 4/13/16.
 */
public class PrivateLocalClient implements LocalClient {
    private final long ACK_TIMEOUT = 2000;
    private final Logger log = Logger.getLogger(PrivateLocalClient.class.getName());

    private Channel comSatChannel;
    private ByteArrayOutputStream buffer;
    private Object ackLock = new Object();
    /**
     * sends buffered messages to ComSat server
     */
    private Runnable sendMessages = () -> {
        IOMessage message;
        while (true) {
            try {
                if (!comSatChannel.isActive())
                    break;

                while ((message = MessageRepository.peekMessage()) == null) {
                    synchronized (MessageRepository.messageLock) {
                        MessageRepository.messageLock.wait();
                    }
                }

                log.info("sending iomessage to comsat");
                comSatChannel.writeAndFlush(message.getBytes()).sync();
                synchronized (ackLock) {
                    try {
                        comSatChannel.writeAndFlush(Constants.TXEND).sync();
                        log.info("waiting for ACK");
                        ackLock.wait(ACK_TIMEOUT);
                        MessageRepository.removeHead();
                    } catch (Exception e) {
                        Thread.sleep(500);
                    }
                }
            } catch (Exception e) {
            }
        }
    };

    public PrivateLocalClient(Channel comSatChannel) {
        buffer = new ByteArrayOutputStream();
        this.comSatChannel = comSatChannel;
        new Thread(sendMessages).start();
    }

    /**
     * handle received data from ComSat server
     *
     * @param message
     */
    @Override
    public void handleMessage(byte[] message) {
        if (message.length < 6) {
            String messageString = new String(message);
            if (messageString.equals("TXEND")) {
                try {
                    log.info("TXEND received");
                    IOMessage ioMessage = new IOMessage(buffer.toByteArray());
                    log.info("sending message to websocket");
                    CoreNetworking.ioFogClient.sendMessageToWebSocket(ioMessage);
                    buffer.reset();
                    log.info("sending ACK to comsat");
                    comSatChannel.writeAndFlush(Constants.ACK).sync();
                } catch (Exception e) {
                }
                return;
            } else if (messageString.equals("ACK")) {
                try {
                    synchronized (ackLock) {
                        ackLock.notifyAll();
                    }
                    log.info("ACK received from comsat");
                    return;
                } catch (Exception e) {
                }
            }
        }
        try {
            log.info("buffering bytes");
            buffer.write(message);
        } catch (Exception e) {
        }
    }

    @Override
    public boolean connect(long timeout) {
        return true;
    }

    @Override
    public boolean isConnected() {
        return true;
    }

    @Override
    public void closeConnection() {

    }
}
