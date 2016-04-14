package main.com.iotracks.core_networking.local_client.private_client;

import com.iotracks.elements.IOMessage;
import io.netty.channel.Channel;
import io.netty.util.CharsetUtil;
import main.com.iotracks.core_networking.main.CoreNetworking;
import main.com.iotracks.core_networking.utils.MessageRepository;
import main.com.iotracks.core_networking.local_client.LocalClient;

import java.io.ByteArrayOutputStream;

/**
 * Created by saeid on 4/13/16.
 */
public class PrivateLocalClient implements LocalClient {
    private final byte[] ACK = "ACK".getBytes(CharsetUtil.US_ASCII);
    private final byte[] TXEND = "TXEND".getBytes(CharsetUtil.US_ASCII);
    private final long ACK_TIMEOUT = 2000;

    private Channel comSatChannel;
    private ByteArrayOutputStream buffer;
    private Object ackLock = new Object();

    public PrivateLocalClient(Channel comSatChannel) {
        buffer = new ByteArrayOutputStream();
        this.comSatChannel = comSatChannel;
        new Thread(sendMessages).start();
    }

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

                comSatChannel.writeAndFlush(message.getBytes()).sync();
                synchronized (ackLock) {
                    try {
                        comSatChannel.writeAndFlush(TXEND).sync();
                        ackLock.wait(ACK_TIMEOUT);
                        MessageRepository.removeHead();
                    } catch (Exception e) {
                        Thread.sleep(500);
                    }
                }
            } catch (Exception e) {}
        }
    };

    @Override
    public void sendMessage(byte[] message) {
        if (message.length < 6) {
            String messageString = new String(message);
            if (messageString.equals("TXEND")) {
                try {
                    IOMessage ioMessage = new IOMessage(buffer.toByteArray());
                    CoreNetworking.ioFabricClient.sendMessageToWebSocket(ioMessage);
                    buffer.reset();
                    comSatChannel.writeAndFlush(ACK).sync();
                } catch (Exception e) {}
                return;
            } else if (messageString.equals("ACK")) {
                synchronized (ackLock) {
                    ackLock.notifyAll();
                }
                return;
            }
        }
        try {
            buffer.write(message);
        } catch (Exception e) {}
    }

    @Override
    public boolean connect(long timeout) {
        return true;
    }

    @Override
    public boolean isConnected() {
        return true;
    }
}
